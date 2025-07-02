# 🚀 SkTorrent Crawler & Search

Efektivní crawler a vyhledávací systém pro torrenty ze SkTorrent.eu s ukládáním do SQLite databáze.

## ✨ Funkce

- 🔍 **Paralelní crawling** s konfigurovatelným počtem workerů
- 💾 **SQLite databáze** s FTS5 full-text search
- ⭐ **ČSFD integrace** - automatické stahování hodnocení a URL
- 🔄 **Automatický update** - při opětovném crawlingu se data aktualizují
- 🖼️ **Metadata extrakce** - obrázky, kategorie, seeders, leechers
- 🔎 **Rychlé vyhledávání** podle názvu nebo kategorie

## 🛠️ Instalace

```bash
# Naklonování repozitáře
git clone <repo-url>
cd search-me-plz-sktorrent

# Instalace závislostí
go mod tidy

# Sestavení aplikací
go build -o crawler cmd/app/main.go
go build -o search cmd/search/main.go
```

## 🚀 Použití

### Crawler - Stahování torrentů

```bash
# Stáhnout stránky 0-5 se 3 workery
./crawler -from=0 -to=5 -workers=3

# Základní parametry
./crawler -from=0 -to=10 -workers=5 -timeout=60 -db=my_torrents.db
```

**Parametry:**
- `-from=N` - Počáteční stránka (default: 0)
- `-to=N` - Koncová stránka (default: 2)
- `-workers=N` - Počet workerů (default: 3, max: 20)
- `-timeout=N` - HTTP timeout v sekundách (default: 30)
- `-db=path` - Cesta k SQLite databázi (default: torrents.db)

### Search - Vyhledávání

```bash
# Vyhledání podle názvu
./search -q "john wick"
./search -q "batman"

# Filtrování podle kategorie
./search -category "Filmy CZ/SK dabing"
./search -category "Seriál"

# Nejnovější torrenty
./search -recent -limit=10

# Statistiky databáze
./search -stats

# Kombinace parametrů
./search -q "avengers" -limit=5 -db=custom.db
```

**Parametry:**
- `-q "text"` - Vyhledávání podle názvu
- `-category "typ"` - Filtrování podle kategorie
- `-recent` - Nejnovější torrenty
- `-stats` - Statistiky databáze
- `-limit=N` - Počet výsledků (default: 20)
- `-db=path` - Cesta k databázi (default: torrents.db)

## 📊 Příklad výstupu

### Crawler
```
🚀 Spouštím SkTorrent Crawler
📄 Stránky: 0 - 2
⚙️  Workery: 3
🗃️  Databáze: torrents.db
✅ Databáze inicializována

💾 UKLÁDÁNÍ STRÁNKY 0 - 24 TORRENTŮ
  ✅ Ironheart S01E04-E06 = CSFD 39%
  ✅ Noční můra v Elm Street 1 = CSFD 75%
  ...

🎉 CRAWLING DOKONČEN!
📊 Celkový počet torrentů: 72
💾 Uloženo do databáze: 72

📈 STATISTIKY DATABÁZE:
  🗃️  Celkem torrentů: 72
  📁 Filmy CZ/SK dabing: 25
  📁 Seriál: 20
  📁 TV Pořad: 15
```

### Search
```
🔍 Vyhledávám: "batman"
✅ Nalezeno 3 výsledků:

[1] 📺 Batman Begins (2005) CSFD 85%
    🆔 ID: abc123...
    🏷️  Kategorie: Filmy CZ/SK dabing
    📦 Velikost: 4.2 GB
    🌱 Seeders: 45 | 🩸 Leechers: 8
    ⭐ ČSFD: 85% (https://www.csfd.cz/film/...)
    🖼️  Obrázek: https://cdn.sktorrent.eu/...
    🔗 URL: https://sktorrent.eu/torrent/details...
    📅 Přidáno: 02.07.2025 09:30
```

## 🗄️ Databázová struktura

```sql
-- Hlavní tabulka torrentů
CREATE TABLE torrents (
    id TEXT PRIMARY KEY,           -- Unikátní hash ID
    name TEXT NOT NULL,            -- Název torrentu
    category TEXT,                 -- Kategorie (Filmy, Seriál, ...)
    size TEXT,                     -- Velikost souboru
    seeds INTEGER,                 -- Počet seeders
    leeches INTEGER,               -- Počet leechers
    url TEXT,                      -- URL na detail stránku
    image_url TEXT,                -- URL náhledu
    csfd_rating TEXT,              -- ČSFD hodnocení (77%)
    csfd_url TEXT,                 -- URL na ČSFD
    created_at DATETIME,           -- Datum prvního přidání
    updated_at DATETIME            -- Datum posledního update
);

-- FTS5 index pro rychlé vyhledávání
CREATE VIRTUAL TABLE torrents_fts USING fts5(
    name, category, content='torrents'
);
```

## 🔧 Architektura

```
cmd/
├── app/main.go       # Crawler aplikace
├── search/main.go    # Search aplikace

internal/
├── crawler/          # Crawling logika
│   └── crawler.go
├── database/         # SQLite databáze
│   └── database.go
```

## 🚨 UPSERT funkcionalita

Aplikace automaticky:
- **Přidá nové torrenty** do databáze
- **Aktualizuje existující** na základě ID
- **Neztratí historii** - uchovává created_at
- **Aktualizuje metadata** - seeders, leechers, ČSFD

## 🎯 Kategorie torrentů

- **Filmy CZ/SK dabing**
- **Filmy s titulkama**
- **HD Filmy**
- **Seriál**
- **TV Pořad**
- **Sport**
- **Hry na Windows**
- **Knihy a Časopisy**

## ⚡ Performance

- **FTS5 full-text search** pro rychlé vyhledávání
- **Paralelní processing** až 20 workerů
- **ČSFD cache** - URL se stahují jen jednou
- **SQLite** optimalizace pro read-heavy workload

## 🛡️ Error handling

- Graceful handling HTTP chyb
- Retry mechanismus pro ČSFD
- Validace dat před uložením
- Podrobné error reporting

---

💡 **Tip:** Pro nejlepší výsledky spouštějte crawler pravidelně (např. každou hodinu) pro aktuální data!