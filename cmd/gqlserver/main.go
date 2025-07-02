package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/JaLe29/search-me-plz-sktorrent/graphql"
	"github.com/JaLe29/search-me-plz-sktorrent/internal/database"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("🔧 Connecting to database...")
	db, err := database.NewDatabase("torrents.db")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("⚠️ Error closing database: %v", err)
		}
	}()
	log.Printf("✅ Database connected successfully")

	resolver := &graphql.Resolver{DB: db}

	// Vytvoření GraphQL serveru s výchozí konfigurací (introspection povolena)
	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: resolver}))

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			log.Printf("📡 %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}

	// Nastavení routů
	http.Handle("/", playground.Handler("SkTorrent GraphQL Playground", "/query"))
	http.Handle("/query", corsMiddleware(srv))

	log.Printf("🚀 SkTorrent GraphQL server starting on port %s", port)
	log.Printf("📖 GraphQL Playground: http://localhost:%s/", port)
	log.Printf("🔗 GraphQL Endpoint: http://localhost:%s/query", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
