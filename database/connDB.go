package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // Import PostgreSQL driver
)

// DB is a global variable for database connection
var DB *sql.DB

// InitializeDB initializes the database connection
func InitializeDB() {
	var err error
	DB, err = sql.Open("postgres", "postgres://postgres:postgres@192.168.0.235:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	// Test the connection
	if err := DB.Ping(); err != nil {
		log.Fatal("Database is unreachable:", err)
	}

	log.Println("Connected to PostgreSQL successfully!")
}
