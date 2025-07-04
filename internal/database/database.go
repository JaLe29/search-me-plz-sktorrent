package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Torrent struct {
	ID         string
	Name       string
	Category   string
	SizeMB     float64   // velikost v MB (normalizovaná)
	AddedDate  time.Time // datum přidání torrentu na web
	URL        string
	ImageURL   string
	CSFDRating int // hodnocení jako číslo (77 místo "77%")
	CSFDURL    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type TorrentStats struct {
	ID         int    // auto-increment ID pro stats záznam
	TorrentID  string // odkaz na torrent
	Seeds      int
	Leeches    int
	RecordedAt time.Time // kdy byly stats zaznamenány
}

type TorrentWithStats struct {
	Torrent
	Seeds   int // aktuální hodnoty
	Leeches int
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("creating tables: %w", err)
	}

	return database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) createTables() error {
	// Hlavní tabulka torrentů (bez seeds/leeches)
	torrentSchema := `
	CREATE TABLE IF NOT EXISTS torrents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		category TEXT,
		size_mb REAL NOT NULL,
		added_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		url TEXT,
		image_url TEXT,
		csfd_rating INTEGER,
		csfd_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(torrentSchema); err != nil {
		return fmt.Errorf("creating torrents table: %w", err)
	}

	// Tabulka pro sledování seeds/leeches v čase
	statsSchema := `
	CREATE TABLE IF NOT EXISTS torrent_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		torrent_id TEXT NOT NULL,
		seeds INTEGER NOT NULL,
		leeches INTEGER NOT NULL,
		recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (torrent_id) REFERENCES torrents(id),
		UNIQUE(torrent_id, recorded_at) ON CONFLICT REPLACE
	);`

	if _, err := d.db.Exec(statsSchema); err != nil {
		return fmt.Errorf("creating torrent_stats table: %w", err)
	}

	// Index pro rychlejší dotazy na stats
	statsIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_stats_torrent_id ON torrent_stats(torrent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_stats_recorded_at ON torrent_stats(recorded_at);`,
		`CREATE INDEX IF NOT EXISTS idx_stats_torrent_recorded ON torrent_stats(torrent_id, recorded_at DESC);`,
	}

	for _, indexSQL := range statsIndexes {
		if _, err := d.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("creating stats index: %w", err)
		}
	}

	// FTS5 virtual table pro rychlé vyhledávání (beze změny)
	ftsSchema := `
	CREATE VIRTUAL TABLE IF NOT EXISTS torrents_fts USING fts5(
		name,
		category,
		content='torrents',
		content_rowid='rowid'
	);`

	if _, err := d.db.Exec(ftsSchema); err != nil {
		return fmt.Errorf("creating FTS table: %w", err)
	}

	// Triggery pro automatickou synchronizaci FTS
	triggers := []string{
		`CREATE TRIGGER IF NOT EXISTS torrents_insert_fts AFTER INSERT ON torrents BEGIN
			INSERT INTO torrents_fts(rowid, name, category) VALUES (new.rowid, new.name, new.category);
		END;`,

		`CREATE TRIGGER IF NOT EXISTS torrents_update_fts AFTER UPDATE ON torrents BEGIN
			UPDATE torrents_fts SET name = new.name, category = new.category WHERE rowid = new.rowid;
		END;`,

		`CREATE TRIGGER IF NOT EXISTS torrents_delete_fts AFTER DELETE ON torrents BEGIN
			DELETE FROM torrents_fts WHERE rowid = old.rowid;
		END;`,
	}

	for _, trigger := range triggers {
		if _, err := d.db.Exec(trigger); err != nil {
			return fmt.Errorf("creating trigger: %w", err)
		}
	}

	return nil
}

// UpsertTorrent vloží nový torrent nebo aktualizuje existující
func (d *Database) UpsertTorrent(t *Torrent) error {
	query := `
	INSERT INTO torrents (
		id, name, category, size_mb, added_date, url,
		image_url, csfd_rating, csfd_url, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		category = excluded.category,
		size_mb = excluded.size_mb,
		added_date = excluded.added_date,
		url = excluded.url,
		image_url = excluded.image_url,
		csfd_rating = excluded.csfd_rating,
		csfd_url = excluded.csfd_url,
		updated_at = excluded.updated_at
	`

	now := time.Now()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now

	_, err := d.db.Exec(query,
		t.ID, t.Name, t.Category, t.SizeMB, t.AddedDate, t.URL,
		t.ImageURL, t.CSFDRating, t.CSFDURL,
		t.CreatedAt, t.UpdatedAt,
	)

	return err
}

// RecordTorrentStats zaznamená aktuální seeds/leeches pro torrent
func (d *Database) RecordTorrentStats(torrentID string, seeds, leeches int) error {
	query := `
	INSERT INTO torrent_stats (torrent_id, seeds, leeches, recorded_at)
	VALUES (?, ?, ?, ?)
	`

	_, err := d.db.Exec(query, torrentID, seeds, leeches, time.Now())
	return err
}

// GetTorrentWithCurrentStats vrátí torrent s nejnovějšími stats
func (d *Database) GetTorrentWithCurrentStats(torrentID string) (*TorrentWithStats, error) {
	query := `
	SELECT t.id, t.name, t.category, t.size_mb, t.added_date, t.url, t.image_url,
		   t.csfd_rating, t.csfd_url, t.created_at, t.updated_at,
		   COALESCE(s.seeds, 0) as seeds, COALESCE(s.leeches, 0) as leeches
	FROM torrents t
	LEFT JOIN (
		SELECT torrent_id, seeds, leeches,
			   ROW_NUMBER() OVER (PARTITION BY torrent_id ORDER BY recorded_at DESC) as rn
		FROM torrent_stats
	) s ON t.id = s.torrent_id AND s.rn = 1
	WHERE t.id = ?
	`

	var result TorrentWithStats
	err := d.db.QueryRow(query, torrentID).Scan(
		&result.ID, &result.Name, &result.Category, &result.SizeMB, &result.AddedDate,
		&result.URL, &result.ImageURL, &result.CSFDRating, &result.CSFDURL,
		&result.CreatedAt, &result.UpdatedAt, &result.Seeds, &result.Leeches,
	)

	if err != nil {
		return nil, fmt.Errorf("getting torrent with stats: %w", err)
	}

	return &result, nil
}

// SearchTorrents vyhledá torrenty podle názvu nebo kategorie s aktuálními stats
func (d *Database) SearchTorrents(query string, limit int) ([]TorrentWithStats, error) {
	if limit <= 0 {
		limit = 50
	}

	sqlQuery := `
	SELECT t.id, t.name, t.category, t.size_mb, t.added_date, t.url, t.image_url,
		   t.csfd_rating, t.csfd_url, t.created_at, t.updated_at,
		   COALESCE(s.seeds, 0) as seeds, COALESCE(s.leeches, 0) as leeches
	FROM torrents t
	LEFT JOIN (
		SELECT torrent_id, seeds, leeches,
			   ROW_NUMBER() OVER (PARTITION BY torrent_id ORDER BY recorded_at DESC) as rn
		FROM torrent_stats
	) s ON t.id = s.torrent_id AND s.rn = 1
	WHERE (t.name LIKE ? OR t.category LIKE ?)
	ORDER BY t.updated_at DESC
	LIMIT ?
	`

	searchTerm := "%" + query + "%"
	rows, err := d.db.Query(sqlQuery, searchTerm, searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("searching torrents: %w", err)
	}
	defer rows.Close()

	var torrents []TorrentWithStats
	for rows.Next() {
		var t TorrentWithStats
		err := rows.Scan(
			&t.ID, &t.Name, &t.Category, &t.SizeMB, &t.AddedDate,
			&t.URL, &t.ImageURL, &t.CSFDRating, &t.CSFDURL,
			&t.CreatedAt, &t.UpdatedAt, &t.Seeds, &t.Leeches,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning torrent: %w", err)
		}
		torrents = append(torrents, t)
	}

	return torrents, nil
}

// GetTorrentsByCategory vrátí torrenty podle kategorie s aktuálními stats
func (d *Database) GetTorrentsByCategory(category string, limit int) ([]TorrentWithStats, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT t.id, t.name, t.category, t.size_mb, t.added_date, t.url, t.image_url,
		   t.csfd_rating, t.csfd_url, t.created_at, t.updated_at,
		   COALESCE(s.seeds, 0) as seeds, COALESCE(s.leeches, 0) as leeches
	FROM torrents t
	LEFT JOIN (
		SELECT torrent_id, seeds, leeches,
			   ROW_NUMBER() OVER (PARTITION BY torrent_id ORDER BY recorded_at DESC) as rn
		FROM torrent_stats
	) s ON t.id = s.torrent_id AND s.rn = 1
	WHERE t.category = ?
	ORDER BY t.updated_at DESC
	LIMIT ?
	`

	rows, err := d.db.Query(query, category, limit)
	if err != nil {
		return nil, fmt.Errorf("getting torrents by category: %w", err)
	}
	defer rows.Close()

	var torrents []TorrentWithStats
	for rows.Next() {
		var t TorrentWithStats
		err := rows.Scan(
			&t.ID, &t.Name, &t.Category, &t.SizeMB, &t.AddedDate,
			&t.URL, &t.ImageURL, &t.CSFDRating, &t.CSFDURL,
			&t.CreatedAt, &t.UpdatedAt, &t.Seeds, &t.Leeches,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning torrent: %w", err)
		}
		torrents = append(torrents, t)
	}

	return torrents, nil
}

// GetRecentTorrents vrátí nejnovější torrenty s aktuálními stats
func (d *Database) GetRecentTorrents(limit int) ([]TorrentWithStats, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT t.id, t.name, t.category, t.size_mb, t.added_date, t.url, t.image_url,
		   t.csfd_rating, t.csfd_url, t.created_at, t.updated_at,
		   COALESCE(s.seeds, 0) as seeds, COALESCE(s.leeches, 0) as leeches
	FROM torrents t
	LEFT JOIN (
		SELECT torrent_id, seeds, leeches,
			   ROW_NUMBER() OVER (PARTITION BY torrent_id ORDER BY recorded_at DESC) as rn
		FROM torrent_stats
	) s ON t.id = s.torrent_id AND s.rn = 1
	ORDER BY t.updated_at DESC
	LIMIT ?
	`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("getting recent torrents: %w", err)
	}
	defer rows.Close()

	var torrents []TorrentWithStats
	for rows.Next() {
		var t TorrentWithStats
		err := rows.Scan(
			&t.ID, &t.Name, &t.Category, &t.SizeMB, &t.AddedDate,
			&t.URL, &t.ImageURL, &t.CSFDRating, &t.CSFDURL,
			&t.CreatedAt, &t.UpdatedAt, &t.Seeds, &t.Leeches,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning torrent: %w", err)
		}
		torrents = append(torrents, t)
	}

	return torrents, nil
}

// GetTorrentStatsHistory vrátí historii seeds/leeches pro torrent
func (d *Database) GetTorrentStatsHistory(torrentID string, limit int) ([]TorrentStats, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
	SELECT id, torrent_id, seeds, leeches, recorded_at
	FROM torrent_stats
	WHERE torrent_id = ?
	ORDER BY recorded_at DESC
	LIMIT ?
	`

	rows, err := d.db.Query(query, torrentID, limit)
	if err != nil {
		return nil, fmt.Errorf("getting torrent stats history: %w", err)
	}
	defer rows.Close()

	var stats []TorrentStats
	for rows.Next() {
		var s TorrentStats
		err := rows.Scan(&s.ID, &s.TorrentID, &s.Seeds, &s.Leeches, &s.RecordedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning torrent stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetStats vrátí statistiky databáze
func (d *Database) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	// Celkový počet torrentů
	var total int
	err := d.db.QueryRow("SELECT COUNT(*) FROM torrents").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("getting total count: %w", err)
	}
	stats["total"] = total

	// Počet podle kategorií
	rows, err := d.db.Query("SELECT category, COUNT(*) FROM torrents GROUP BY category")
	if err != nil {
		return nil, fmt.Errorf("getting category stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("scanning category stat: %w", err)
		}
		stats[category] = count
	}

	// Počet stats záznamů
	var statsCount int
	err = d.db.QueryRow("SELECT COUNT(*) FROM torrent_stats").Scan(&statsCount)
	if err == nil {
		stats["stats_records"] = statsCount
	}

	return stats, nil
}

// GetTorrentsByCSFDID vrátí torrenty podle CSFD ID s aktuálními stats
func (d *Database) GetTorrentsByCSFDID(csfdID string, limit int) ([]TorrentWithStats, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT t.id, t.name, t.category, t.size_mb, t.added_date, t.url, t.image_url,
		   t.csfd_rating, t.csfd_url, t.created_at, t.updated_at,
		   COALESCE(s.seeds, 0) as seeds, COALESCE(s.leeches, 0) as leeches
	FROM torrents t
	LEFT JOIN (
		SELECT torrent_id, seeds, leeches,
			   ROW_NUMBER() OVER (PARTITION BY torrent_id ORDER BY recorded_at DESC) as rn
		FROM torrent_stats
	) s ON t.id = s.torrent_id AND s.rn = 1
	WHERE t.csfd_url LIKE ?
	ORDER BY t.updated_at DESC
	LIMIT ?
	`

	rows, err := d.db.Query(query, "%"+csfdID+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("getting torrents by CSFD ID: %w", err)
	}
	defer rows.Close()

	var torrents []TorrentWithStats
	for rows.Next() {
		var t TorrentWithStats
		err := rows.Scan(
			&t.ID, &t.Name, &t.Category, &t.SizeMB, &t.AddedDate,
			&t.URL, &t.ImageURL, &t.CSFDRating, &t.CSFDURL,
			&t.CreatedAt, &t.UpdatedAt, &t.Seeds, &t.Leeches,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning torrent: %w", err)
		}
		torrents = append(torrents, t)
	}

	return torrents, nil
}

// GetTorrentsWithPagination vrátí torrenty s stránkováním a informací o dalších stránkách
func (d *Database) GetTorrentsWithPagination(offset, limit int, category *string, search *string, sortBy string) ([]TorrentWithStats, int, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Base query
	baseQuery := `
	SELECT t.id, t.name, t.category, t.size_mb, t.added_date, t.url, t.image_url,
		   t.csfd_rating, t.csfd_url, t.created_at, t.updated_at,
		   COALESCE(s.seeds, 0) as seeds, COALESCE(s.leeches, 0) as leeches
	FROM torrents t
	LEFT JOIN (
		SELECT torrent_id, seeds, leeches,
			   ROW_NUMBER() OVER (PARTITION BY torrent_id ORDER BY recorded_at DESC) as rn
		FROM torrent_stats
	) s ON t.id = s.torrent_id AND s.rn = 1
	`

	// Build WHERE clause
	var whereClause string
	var args []interface{}

	if search != nil && *search != "" {
		// Use LIKE for contains search instead of FTS5 for better partial matching
		whereClause = "WHERE (t.name LIKE ? OR t.category LIKE ?)"
		searchTerm := "%" + *search + "%"
		args = append(args, searchTerm, searchTerm)
	} else if category != nil && *category != "" {
		whereClause = "WHERE t.category = ?"
		args = append(args, *category)
	}

	// Build ORDER BY clause
	var orderBy string
	switch sortBy {
	case "OLDEST":
		orderBy = "ORDER BY t.updated_at ASC"
	case "NAME_ASC":
		orderBy = "ORDER BY t.name ASC"
	case "NAME_DESC":
		orderBy = "ORDER BY t.name DESC"
	case "SIZE_ASC":
		orderBy = "ORDER BY t.size_mb ASC"
	case "SIZE_DESC":
		orderBy = "ORDER BY t.size_mb DESC"
	case "SEEDS_DESC":
		orderBy = "ORDER BY s.seeds DESC"
	case "LEECHES_DESC":
		orderBy = "ORDER BY s.leeches DESC"
	default: // NEWEST
		orderBy = "ORDER BY t.updated_at DESC"
	}

	// Count total results for hasNextPage
	var countQuery string
	if whereClause != "" {
		countQuery = "SELECT COUNT(*) FROM torrents t " + whereClause
	} else {
		countQuery = "SELECT COUNT(*) FROM torrents t"
	}
	var totalCount int
	err := d.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, false, fmt.Errorf("counting torrents: %w", err)
	}

	// Get paginated results
	var query string
	if whereClause != "" {
		query = baseQuery + " " + whereClause + " " + orderBy + " LIMIT ? OFFSET ?"
	} else {
		query = baseQuery + " " + orderBy + " LIMIT ? OFFSET ?"
	}
	args = append(args, limit+1, offset) // +1 to check if there are more results

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, 0, false, fmt.Errorf("getting torrents with pagination: %w", err)
	}
	defer rows.Close()

	var torrents []TorrentWithStats
	for rows.Next() {
		var t TorrentWithStats
		err := rows.Scan(
			&t.ID, &t.Name, &t.Category, &t.SizeMB, &t.AddedDate,
			&t.URL, &t.ImageURL, &t.CSFDRating, &t.CSFDURL,
			&t.CreatedAt, &t.UpdatedAt, &t.Seeds, &t.Leeches,
		)
		if err != nil {
			return nil, 0, false, fmt.Errorf("scanning torrent: %w", err)
		}
		torrents = append(torrents, t)
	}

	// Check if there are more results
	hasNextPage := len(torrents) > limit
	if hasNextPage {
		torrents = torrents[:limit] // Remove the extra item
	}

	return torrents, totalCount, hasNextPage, nil
}
