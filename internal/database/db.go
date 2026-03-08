package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// DB is a global variable holding our connection pool.
// Other files in our project will use this to run SQL queries.
var DB *pgxpool.Pool

// InitDB loads the .env file and connects to Neon PostgreSQL
func InitDB() {
	// 1. Load the .env file (just like your snippet!)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Relying on system environment variables.")
	}

	// 2. Get the connection string
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set in the environment")
	}

	// 3. Connect using a Connection Pool instead of a single connection
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// 4. Ping to verify the connection works
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	fmt.Println("Successfully connected to Neon PostgreSQL!")
	
	// Assign the pool to our global variable
	DB = pool
}

// CloseDB gracefully shuts down the pool when the server stops
func CloseDB() {
	if DB != nil {
		DB.Close()
		fmt.Println("Database connection closed.")
	}
}