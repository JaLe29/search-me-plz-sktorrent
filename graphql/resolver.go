package graphql

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import "github.com/JaLe29/search-me-plz-sktorrent/internal/database"

type Resolver struct {
	DB *database.Database
}
