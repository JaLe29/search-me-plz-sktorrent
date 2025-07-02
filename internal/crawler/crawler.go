package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Torrent struct {
	Name     string
	Category string
	Size     string
	Seeds    int
	Leeches  int
	URL      string
	ImageURL string
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
}

type Crawler struct {
	client *http.Client
	config Config
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

	return &Crawler{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
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

		// URL
		if href, exists := detailsLink.Attr("href"); exists {
			torrent.URL = "https://sktorrent.eu/torrent/" + href
		}

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

		torrents = append(torrents, torrent)
	})

	return torrents
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
				torrent.Size = strings.TrimSpace(strings.Replace(line, "Velkost", "", 1))
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

func (c *Crawler) processResults(results <-chan CrawlResult) {
	resultMap := make(map[int]CrawlResult)

	// Sbírání všech výsledků
	for result := range results {
		resultMap[result.PageNum] = result
	}

	// Zobrazení výsledků v pořadí
	totalTorrents := 0
	for pageNum := range resultMap {
		result := resultMap[pageNum]

		if result.Error != nil {
			fmt.Printf("❌ CHYBA na stránce %d: %v\n", result.PageNum, result.Error)
			continue
		}

		fmt.Printf("\n🔥 STRÁNKA %d - NALEZENO %d TORRENTŮ 🔥\n", result.PageNum, len(result.Torrents))

		for j, torrent := range result.Torrents {
			fmt.Printf("[%d] 📺 %s\n", j+1, torrent.Name)
			fmt.Printf("    🏷️  Kategorie: %s\n", torrent.Category)
			fmt.Printf("    📦 Velikost: %s\n", torrent.Size)
			fmt.Printf("    🌱 Seeders: %d | 🩸 Leechers: %d\n", torrent.Seeds, torrent.Leeches)
			if torrent.ImageURL != "" {
				fmt.Printf("    🖼️  Obrázek: %s\n", torrent.ImageURL)
			}
			fmt.Printf("    🔗 URL: %s\n", torrent.URL)
			fmt.Println("    " + strings.Repeat("─", 60))
		}

		totalTorrents += len(result.Torrents)
		fmt.Printf("✅ KONEC STRÁNKY %d\n", result.PageNum)
	}

	fmt.Printf("\n🎉 CRAWLING DOKONČEN! 🎉\n")
	fmt.Printf("📊 Celkový počet torrentů: %d\n", totalTorrents)
	fmt.Printf("⚙️  Použito workerů: %d\n", c.config.Workers)
}
