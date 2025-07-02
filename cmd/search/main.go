package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/JaLe29/search-me-plz-sktorrent/internal/database"
)

func main() {
	// Definice příkazových parametrů
	var (
		dbPath       = flag.String("db", "torrents.db", "Cesta k SQLite databázi")
		query        = flag.String("q", "", "Vyhledávací dotaz (název nebo kategorie)")
		category     = flag.String("category", "", "Filtrovat podle kategorie")
		recent       = flag.Bool("recent", false, "Zobrazit nejnovější torrenty")
		limit        = flag.Int("limit", 20, "Maximální počet výsledků")
		stats        = flag.Bool("stats", false, "Zobrazit statistiky databáze")
		history      = flag.String("history", "", "Zobrazit historii stats pro torrent ID")
		historyLimit = flag.Int("history-limit", 50, "Počet historických záznamů")
	)
	flag.Parse()

	if *query == "" && *category == "" && !*recent && !*stats && *history == "" {
		fmt.Println("🔍 SkTorrent Search")
		fmt.Println("Použití:")
		fmt.Println("  -q \"text\"          Vyhledat podle názvu")
		fmt.Println("  -category \"typ\"    Filtrovat podle kategorie")
		fmt.Println("  -recent            Zobrazit nejnovější")
		fmt.Println("  -stats             Zobrazit statistiky")
		fmt.Println("  -history \"id\"      Zobrazit historii stats pro torrent")
		fmt.Println("  -limit N           Počet výsledků (default: 20)")
		fmt.Println("  -history-limit N   Počet historických záznamů (default: 50)")
		fmt.Println("  -db path           Cesta k databázi (default: torrents.db)")
		fmt.Println()
		fmt.Println("Příklady:")
		fmt.Println("  ./search -q \"john wick\"")
		fmt.Println("  ./search -category \"Filmy CZ/SK dabing\"")
		fmt.Println("  ./search -recent")
		fmt.Println("  ./search -stats")
		fmt.Println("  ./search -history \"abc123...\"")
		return
	}

	// Inicializace databáze
	db, err := database.NewDatabase(*dbPath)
	if err != nil {
		log.Fatalf("❌ Chyba při připojení k databázi: %v", err)
	}
	defer db.Close()

	// Zobrazení statistik
	if *stats {
		showStats(db)
		return
	}

	// Zobrazení historie stats
	if *history != "" {
		showStatsHistory(db, *history, *historyLimit)
		return
	}

	var torrents []database.TorrentWithStats

	// Vyhledávání podle parametrů
	if *query != "" {
		fmt.Printf("🔍 Vyhledávám: \"%s\"\n", *query)
		torrents, err = db.SearchTorrents(*query, *limit)
	} else if *category != "" {
		fmt.Printf("📁 Kategorie: \"%s\"\n", *category)
		torrents, err = db.GetTorrentsByCategory(*category, *limit)
	} else if *recent {
		fmt.Printf("⏰ Nejnovější torrenty:\n")
		torrents, err = db.GetRecentTorrents(*limit)
	}

	if err != nil {
		log.Fatalf("❌ Chyba při vyhledávání: %v", err)
	}

	if len(torrents) == 0 {
		fmt.Println("❌ Žádné výsledky nenalezeny")
		return
	}

	fmt.Printf("✅ Nalezeno %d výsledků:\n\n", len(torrents))

	// Zobrazení výsledků
	for i, torrent := range torrents {
		fmt.Printf("[%d] 📺 %s\n", i+1, torrent.Name)
		fmt.Printf("    🆔 ID: %s\n", torrent.ID)
		fmt.Printf("    🏷️  Kategorie: %s\n", torrent.Category)
		fmt.Printf("    📦 Velikost: %.1f MB\n", torrent.SizeMB)
		fmt.Printf("    📅 Přidáno na web: %s\n", torrent.AddedDate.Format("02.01.2006"))
		fmt.Printf("    🌱 Seeders: %d | 🩸 Leechers: %d\n", torrent.Seeds, torrent.Leeches)

		if torrent.CSFDRating != 0 {
			fmt.Printf("    ⭐ ČSFD: %d%%", torrent.CSFDRating)
			if torrent.CSFDURL != "" {
				fmt.Printf(" (%s)", torrent.CSFDURL)
			}
			fmt.Println()
		}

		if torrent.ImageURL != "" {
			fmt.Printf("    🖼️  Obrázek: %s\n", torrent.ImageURL)
		}

		fmt.Printf("    🔗 URL: %s\n", torrent.URL)
		fmt.Printf("    📅 Přidáno do DB: %s\n", torrent.CreatedAt.Format("02.01.2006 15:04"))
		fmt.Printf("    🔄 Aktualizováno: %s\n", torrent.UpdatedAt.Format("02.01.2006 15:04"))
		fmt.Printf("    💡 Historie: ./search -history \"%s\"\n", torrent.ID)
		fmt.Println("    " + strings.Repeat("─", 60))
	}
}

func showStats(db *database.Database) {
	stats, err := db.GetStats()
	if err != nil {
		log.Fatalf("❌ Chyba při získávání statistik: %v", err)
	}

	fmt.Printf("📈 STATISTIKY DATABÁZE\n")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Printf("🗃️  Celkem torrentů: %d\n", stats["total"])
	if statsRecords, ok := stats["stats_records"]; ok {
		fmt.Printf("📊 Stats záznamů: %d\n", statsRecords)
	}
	fmt.Println()

	if len(stats) > 1 {
		fmt.Printf("📁 PODLE KATEGORIÍ:\n")
		for category, count := range stats {
			if category != "total" && category != "stats_records" && count > 0 {
				fmt.Printf("  • %-25s %d\n", category, count)
			}
		}
	}

	// Zobrazení nejnovějších torrentů
	fmt.Printf("\n⏰ NEJNOVĚJŠÍ TORRENTY:\n")
	recent, err := db.GetRecentTorrents(5)
	if err == nil && len(recent) > 0 {
		for i, torrent := range recent {
			fmt.Printf("  %d. %s", i+1, torrent.Name)
			if torrent.CSFDRating != 0 {
				fmt.Printf(" (ČSFD: %d%%)", torrent.CSFDRating)
			}
			fmt.Printf("\n     📅 %s | 🌱 %d 🩸 %d\n",
				torrent.UpdatedAt.Format("02.01.2006 15:04"),
				torrent.Seeds, torrent.Leeches)
		}
	}
}

func showStatsHistory(db *database.Database, torrentID string, limit int) {
	// Nejprve získáme základní info o torrentu
	torrent, err := db.GetTorrentWithCurrentStats(torrentID)
	if err != nil {
		log.Fatalf("❌ Torrent s ID '%s' nenalezen: %v", torrentID, err)
	}

	fmt.Printf("📊 HISTORIE STATS PRO TORRENT\n")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("📺 %s\n", torrent.Name)
	fmt.Printf("🆔 ID: %s\n", torrent.ID)
	fmt.Printf("🏷️  Kategorie: %s\n", torrent.Category)
	fmt.Printf("📦 Velikost: %.1f MB\n", torrent.SizeMB)
	fmt.Printf("📅 Přidáno na web: %s\n", torrent.AddedDate.Format("02.01.2006"))
	fmt.Printf("🌱 Aktuální Seeders: %d | 🩸 Leechers: %d\n\n", torrent.Seeds, torrent.Leeches)

	// Získáme historii
	history, err := db.GetTorrentStatsHistory(torrentID, limit)
	if err != nil {
		log.Fatalf("❌ Chyba při získávání historie: %v", err)
	}

	if len(history) == 0 {
		fmt.Println("❌ Žádná historie stats nenalezena")
		return
	}

	fmt.Printf("📈 HISTORIE (%d záznamů):\n", len(history))
	fmt.Println("┌────────────────────┬─────────┬─────────┐")
	fmt.Println("│ Čas                │ Seeders │ Leechers│")
	fmt.Println("├────────────────────┼─────────┼─────────┤")

	for _, stat := range history {
		fmt.Printf("│ %-18s │ %7d │ %7d │\n",
			stat.RecordedAt.Format("02.01.06 15:04"),
			stat.Seeds,
			stat.Leeches)
	}
	fmt.Println("└────────────────────┴─────────┴─────────┘")

	// Jednoduchá analýza trendu
	if len(history) >= 2 {
		first := history[len(history)-1] // nejstarší
		last := history[0]               // nejnovější

		seedsDiff := last.Seeds - first.Seeds
		leechesDiff := last.Leeches - first.Leeches

		fmt.Printf("\n📊 TREND (od %s):\n", first.RecordedAt.Format("02.01 15:04"))

		if seedsDiff > 0 {
			fmt.Printf("   🔼 Seeders: +%d\n", seedsDiff)
		} else if seedsDiff < 0 {
			fmt.Printf("   🔽 Seeders: %d\n", seedsDiff)
		} else {
			fmt.Printf("   ➡️  Seeders: beze změny\n")
		}

		if leechesDiff > 0 {
			fmt.Printf("   🔼 Leechers: +%d\n", leechesDiff)
		} else if leechesDiff < 0 {
			fmt.Printf("   🔽 Leechers: %d\n", leechesDiff)
		} else {
			fmt.Printf("   ➡️  Leechers: beze změny\n")
		}
	}
}
