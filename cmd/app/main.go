package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JaLe29/search-me-plz-sktorrent/internal/crawler"
	"github.com/JaLe29/search-me-plz-sktorrent/internal/database"
)

func main() {
	// Definice příkazových parametrů
	var (
		fromPage = flag.Int("from", 0, "Počáteční stránka pro crawling (začíná od 0)")
		toPage   = flag.Int("to", 2, "Koncová stránka pro crawling")
		workers  = flag.Int("workers", 3, "Počet paralelních workerů")
		timeout  = flag.Int("timeout", 30, "Timeout pro HTTP požadavky (sekundy)")
		dbPath   = flag.String("db", "torrents.db", "Cesta k SQLite databázi")
	)
	flag.Parse()

	// Validace parametrů
	if *fromPage < 0 || *toPage < 0 || *fromPage > *toPage {
		log.Fatal("❌ Neplatné rozmezí stránek. Použij -from=0 -to=10")
	}
	if *workers < 1 || *workers > 20 {
		log.Fatal("❌ Počet workerů musí být mezi 1-20")
	}

	fmt.Printf("🚀 Spouštím SkTorrent Crawler\n")
	fmt.Printf("📄 Stránky: %d - %d\n", *fromPage, *toPage)
	fmt.Printf("⚙️  Workery: %d\n", *workers)
	fmt.Printf("⏱️  Timeout: %ds\n", *timeout)
	fmt.Printf("🗃️  Databáze: %s\n", *dbPath)
	fmt.Println(strings.Repeat("=", 50))

	// Inicializace databáze
	db, err := database.NewDatabase(*dbPath)
	if err != nil {
		log.Fatalf("❌ Chyba při inicializaci databáze: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("⚠️  Chyba při zavírání databáze: %v\n", err)
		}
	}()

	fmt.Printf("✅ Databáze inicializována\n")

	// Konfigurace crawleru
	config := crawler.Config{
		Workers:  *workers,
		Timeout:  time.Duration(*timeout) * time.Second,
		Database: db,
	}

	// Vytvoření a spuštění crawleru
	c := crawler.NewCrawler(config)

	startTime := time.Now()
	c.Crawl(*fromPage, *toPage)
	duration := time.Since(startTime)

	fmt.Printf("\n⏱️  Celkový čas: %v\n", duration)
	fmt.Println("🎉 Hotovo!")
}
