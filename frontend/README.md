# SkTorrent Frontend

React aplikace pro prohlížení torrentů z SkTorrent databáze pomocí GraphQL API.

## Technologie

- **React 19** s TypeScript
- **Ant Design** pro UI komponenty
- **Apollo Client** pro GraphQL komunikaci
- **Biome** pro linting a formátování
- **Vite** pro build tooling

## Instalace

```bash
npm install
```

## Vývoj

Spuštění vývojového serveru:

```bash
npm run dev
```

Aplikace bude dostupná na `http://localhost:5173`

## Build

Vytvoření produkčního buildu:

```bash
npm run build
```

## Linting a formátování

```bash
# Kontrola kódu
npm run lint

# Formátování kódu
npm run format

# Kompletní kontrola (lint + format)
npm run check
```

## Funkce

- 📋 Seznam torrentů s paginací
- 🔍 Vyhledávání torrentů
- 🏷️ Filtrování podle kategorií
- 📊 Řazení podle různých kritérií
- 🎬 Zobrazení ČSFD hodnocení
- 📱 Responzivní design

## Struktura projektu

```
src/
├── components/          # React komponenty
│   ├── TorrentCard.tsx  # Karta jednotlivého torrentu
│   └── TorrentList.tsx  # Seznam torrentů s filtry
├── graphql/             # GraphQL queries
│   └── queries.ts       # Definice GraphQL dotazů
├── lib/                 # Knihovny a konfigurace
│   └── apollo.ts        # Apollo Client konfigurace
├── types/               # TypeScript typy
│   └── graphql.ts       # Typy pro GraphQL schéma
├── App.tsx              # Hlavní komponenta
└── main.tsx             # Entry point
```

## Konfigurace

GraphQL endpoint je nastaven na `http://localhost:8080/query` v `src/lib/apollo.ts`. Pro změnu endpointu upravte tuto hodnotu.
