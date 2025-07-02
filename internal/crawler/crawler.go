package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JaLe29/search-me-plz-sktorrent/internal/database"
	"github.com/PuerkitoBio/goquery"
)

type Torrent struct {
	ID         string // unikátní ID torrentu z URL
	Name       string
	Category   string
	SizeMB     float64   // velikost v MB
	AddedDate  time.Time // datum přidání
	Seeds      int
	Leeches    int
	URL        string
	ImageURL   string
	CSFDRating int    // hodnocení jako číslo (77 místo "77%")
	CSFDURL    string // přímý odkaz na ČSFD
}

type CrawlResult struct {
	PageNum  int
	Torrents []Torrent
	Error    error
}

type Config struct {
	Workers   int
	BaseURL   string
	UserAgent string
	Timeout   time.Duration
	Database  *database.Database // databáze pro ukládání
}

type Crawler struct {
	client    *http.Client
	config    Config
	csfdRegex *regexp.Regexp
}

func NewCrawler(config Config) *Crawler {
	if config.BaseURL == "" {
		config.BaseURL = "https://sktorrent.eu/torrent/torrents_v2.php?active=0&order=data&by=DESC&zaner=&jazyk=&page=0"
	}
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (compatible; SkTorrent-Crawler/1.0)"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Workers <= 0 {
		config.Workers = 3
	}

	// Regex pro parsování ČSFD hodnocení z názvu
	csfdRegex := regexp.MustCompile(`=\s*CSFD\s*(\d+)%`)

	return &Crawler{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config:    config,
		csfdRegex: csfdRegex,
	}
}

func (c *Crawler) Crawl(from int, to int) {
	// Vytvoření kanálů pro paralelní zpracování
	jobs := make(chan int, to-from+1)
	results := make(chan CrawlResult, to-from+1)

	// Spuštění workerů
	var wg sync.WaitGroup
	for i := 0; i < c.config.Workers; i++ {
		wg.Add(1)
		go c.worker(jobs, results, &wg)
	}

	// Odeslání úkolů do kanálu
	go func() {
		for pageNum := from; pageNum <= to; pageNum++ {
			jobs <- pageNum
		}
		close(jobs)
	}()

	// Čekání na dokončení všech workerů
	go func() {
		wg.Wait()
		close(results)
	}()

	// Sbírání a zobrazení výsledků
	c.processResults(results)
}

func (c *Crawler) worker(jobs <-chan int, results chan<- CrawlResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for pageNum := range jobs {
		fmt.Printf("Worker processing page %d...\n", pageNum)
		torrents, err := c.crawlPage(pageNum)
		results <- CrawlResult{
			PageNum:  pageNum,
			Torrents: torrents,
			Error:    err,
		}
	}
}

func (c *Crawler) crawlPage(pageNum int) ([]Torrent, error) {
	url := fmt.Sprintf("%s&page=%d", c.config.BaseURL, pageNum)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching page %d: %w", pageNum, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("page %d returned status %d", pageNum, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Parsování torrentů přímo z paměti
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	torrents := c.parseTorrents(doc)
	return torrents, nil
}

func (c *Crawler) parseTorrents(doc *goquery.Document) []Torrent {
	var torrents []Torrent

	doc.Find("TD.lista").Each(func(i int, s *goquery.Selection) {
		// Zkontrolovat, zda obsahuje odkaz na details.php
		detailsLink := s.Find("A[href*='details.php']")
		if detailsLink.Length() == 0 {
			return
		}

		var torrent Torrent

		// Název torrentu
		torrent.Name = strings.TrimSpace(detailsLink.Text())
		if torrent.Name == "" {
			return
		}

		// URL a ID
		if href, exists := detailsLink.Attr("href"); exists {
			torrent.URL = "https://sktorrent.eu/torrent/" + href
			// Extrakce ID z URL parametrů
			torrent.ID = c.extractTorrentID(href)
		}

		// ČSFD hodnocení z názvu
		torrent.CSFDRating = c.parseCSFDRating(torrent.Name)

		// Kategorie
		categoryLink := s.Find("a[href*='torrents_v2.php?category=']")
		if categoryLink.Length() > 0 {
			torrent.Category = strings.TrimSpace(categoryLink.Text())
		}

		// URL obrázku
		imgElement := s.Find("img.lozad")
		if imgElement.Length() > 0 {
			if dataSrc, exists := imgElement.Attr("data-src"); exists {
				torrent.ImageURL = dataSrc
			}
		}

		// Velikost, seeders, leechers
		c.parseMetadata(s, &torrent)

		// Vždy stáhnout přímý ČSFD odkaz z detail stránky (pokud má ČSFD hodnocení)
		if torrent.CSFDRating != 0 {
			torrent.CSFDURL = c.fetchCSFDURL(torrent.URL)
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

func (c *Crawler) extractTorrentID(href string) string {
	// href vypadá jako: details.php?name=...&id=339688748bd23e2ec25945937872287be91343f9
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return u.Query().Get("id")
}

func (c *Crawler) parseCSFDRating(name string) int {
	matches := c.csfdRegex.FindStringSubmatch(name)
	if len(matches) >= 2 {
		rating, err := strconv.Atoi(matches[1])
		if err == nil {
			return rating
		}
	}
	return 0
}

func (c *Crawler) fetchCSFDURL(detailURL string) string {
	if detailURL == "" {
		return ""
	}

	req, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return ""
	}

	// Hledat ČSFD odkaz pomocí různých selektorů
	csfdSelectors := []string{
		`a[itemprop="sameAs"][href*="csfd.cz"]`,
		`a[href*="csfd.cz/film/"]`,
		`a[href*="csfd.sk/film/"]`,
	}

	for _, selector := range csfdSelectors {
		link := doc.Find(selector).First()
		if link.Length() > 0 {
			if href, exists := link.Attr("href"); exists {
				return href
			}
		}
	}

	return ""
}

func (c *Crawler) parseMetadata(s *goquery.Selection, torrent *Torrent) {
	s.Find("*").Each(func(j int, textNode *goquery.Selection) {
		text := textNode.Text()
		if !strings.Contains(text, "Velkost") {
			return
		}

		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "Velkost") {
				// Parsovat velikost a datum z řádku jako "Velkost: 6.9 GB | Pridany 02/07/2025"
				c.parseSizeAndDate(line, torrent)
			} else if strings.HasPrefix(line, "Odosielaju") {
				seedText := strings.TrimSpace(strings.Replace(line, "Odosielaju :", "", 1))
				fmt.Sscanf(seedText, "%d", &torrent.Seeds)
			} else if strings.HasPrefix(line, "Stahuju") {
				leechText := strings.TrimSpace(strings.Replace(line, "Stahuju :", "", 1))
				fmt.Sscanf(leechText, "%d", &torrent.Leeches)
			}
		}
	})
}

func (c *Crawler) parseSizeAndDate(line string, torrent *Torrent) {
	// Očekáváme formát: "Velkost 6.9 GB | Pridany 02/07/2025"
	parts := strings.Split(line, "|")

	// Parsování velikosti
	if len(parts) >= 1 {
		sizePart := strings.TrimSpace(parts[0])
		sizePart = strings.Replace(sizePart, "Velkost", "", 1)
		sizePart = strings.TrimSpace(sizePart)
		torrent.SizeMB = c.parseSizeMB(sizePart)
	}

	// Parsování data
	if len(parts) >= 2 {
		datePart := strings.TrimSpace(parts[1])
		datePart = strings.Replace(datePart, "Pridany", "", 1)
		datePart = strings.TrimSpace(datePart)
		torrent.AddedDate = c.parseAddedDate(datePart)
	}
}

func (c *Crawler) parseSizeMB(sizeStr string) float64 {
	// Parsování velikosti z formátu "6.9 GB", "1.2 TB", "500 MB" atd.
	re := regexp.MustCompile(`([0-9.]+)\s*(GB|TB|MB|KB)`)
	matches := re.FindStringSubmatch(sizeStr)

	if len(matches) != 3 {
		return 0.0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.0
	}

	unit := strings.ToUpper(matches[2])
	switch unit {
	case "KB":
		return value / 1024 // Convert KB to MB
	case "MB":
		return value
	case "GB":
		return value * 1024 // Convert GB to MB
	case "TB":
		return value * 1024 * 1024 // Convert TB to MB
	default:
		return value
	}
}

func (c *Crawler) parseAddedDate(dateStr string) time.Time {
	// Parsování data z formátu "02/07/2025"
	layouts := []string{
		"02/01/2006",
		"2/1/2006",
		"02/1/2006",
		"2/01/2006",
	}

	for _, layout := range layouts {
		if parsedTime, err := time.Parse(layout, dateStr); err == nil {
			return parsedTime
		}
	}

	// Pokud se nepodaří parsovat, vrátíme aktuální čas
	return time.Now()
}

func (c *Crawler) processResults(results <-chan CrawlResult) {
	resultMap := make(map[int]CrawlResult)

	// Sbírání všech výsledků
	for result := range results {
		resultMap[result.PageNum] = result
	}

	// Zpracování výsledků v pořadí
	totalTorrents := 0
	savedTorrents := 0

	for pageNum := range resultMap {
		result := resultMap[pageNum]

		if result.Error != nil {
			fmt.Printf("❌ CHYBA na stránce %d: %v\n", result.PageNum, result.Error)
			continue
		}

		fmt.Printf("💾 UKLÁDÁNÍ STRÁNKY %d - %d TORRENTŮ\n", result.PageNum, len(result.Torrents))

		// Uložení torrentů do databáze
		for _, torrent := range result.Torrents {
			if c.config.Database != nil {
				// Uložit základní informace o torrentu
				dbTorrent := c.convertToDBTorrent(torrent)
				if err := c.config.Database.UpsertTorrent(&dbTorrent); err != nil {
					fmt.Printf("⚠️  Chyba při ukládání torrentu %s: %v\n", torrent.ID, err)
					continue
				}

				// Zaznamenat aktuální stats (seeds/leeches) s časovým razítkem
				if err := c.config.Database.RecordTorrentStats(torrent.ID, torrent.Seeds, torrent.Leeches); err != nil {
					fmt.Printf("⚠️  Chyba při ukládání stats pro %s: %v\n", torrent.ID, err)
				}

				savedTorrents++
			}

			// Stručný výpis
			fmt.Printf("  ✅ %s", torrent.Name)
			if torrent.CSFDRating != 0 {
				fmt.Printf(" (ČSFD: %d%%)", torrent.CSFDRating)
			}
			fmt.Printf(" [%.1f MB]", torrent.SizeMB)
			fmt.Printf(" [S:%d L:%d]", torrent.Seeds, torrent.Leeches)
			if !torrent.AddedDate.IsZero() {
				fmt.Printf(" [%s]", torrent.AddedDate.Format("02.01.06"))
			}
			fmt.Printf("\n")
		}

		totalTorrents += len(result.Torrents)
	}

	fmt.Printf("\n🎉 CRAWLING DOKONČEN! 🎉\n")
	fmt.Printf("📊 Celkový počet torrentů: %d\n", totalTorrents)
	fmt.Printf("💾 Uloženo do databáze: %d\n", savedTorrents)
	fmt.Printf("⚙️  Použito workerů: %d\n", c.config.Workers)

	// Zobrazení statistik databáze
	if c.config.Database != nil {
		if stats, err := c.config.Database.GetStats(); err == nil {
			fmt.Printf("\n📈 STATISTIKY DATABÁZE:\n")
			fmt.Printf("  🗃️  Celkem torrentů: %d\n", stats["total"])
			for category, count := range stats {
				if category != "total" && count > 0 {
					fmt.Printf("  📁 %s: %d\n", category, count)
				}
			}
		}
	}
}

// convertToDBTorrent převede crawler.Torrent na database.Torrent
func (c *Crawler) convertToDBTorrent(t Torrent) database.Torrent {
	return database.Torrent{
		ID:         t.ID,
		Name:       t.Name,
		Category:   t.Category,
		SizeMB:     t.SizeMB,
		AddedDate:  t.AddedDate,
		URL:        t.URL,
		ImageURL:   t.ImageURL,
		CSFDRating: t.CSFDRating,
		CSFDURL:    t.CSFDURL,
	}
}
