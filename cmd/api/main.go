package main

import (
	"fmt"
	"log"
	"net/http"

	// Replace this with your actual module path
	"github.com/divyanshmehta355/charcha-backend/internal/cache"
	"github.com/divyanshmehta355/charcha-backend/internal/database"
)

func main() {
	fmt.Println("Starting Charcha Backend...")

	// 1. Initialize our Neon PostgreSQL connection
	database.InitDB()
	// defer ensures the database closes gracefully when the server stops
	defer database.CloseDB()

	// 2. Initialize our Upstash Redis connection
	cache.InitRedis()
	defer cache.CloseRedis()

	// 3. Set up a basic health check route
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Charcha API is running and connected to the cloud!"))
	})

	// 4. Start the HTTP server
	port := ":8080"
	fmt.Printf("Server is listening on http://localhost%s\n", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
