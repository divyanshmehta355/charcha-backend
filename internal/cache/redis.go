package cache

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// Ctx is a global context we will use for all Redis operations
var Ctx = context.Background()

// Client is our global Redis client.
// Other packages will use cache.Client to publish and subscribe to messages.
var Client *redis.Client

// InitRedis connects to Upstash Redis using the connection string
func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is not set")
	}

	// 1. Parse the URL (and catch any errors!)
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Error parsing REDIS_URL: %v", err)
	}

	// 2. Create the new client using the parsed options
	Client = redis.NewClient(opt)

	// 3. Ping the database to verify the connection is active
	_, err = Client.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Upstash Redis: %v", err)
	}

	fmt.Println("Successfully connected to Upstash Redis!")
}

// CloseRedis gracefully closes the connection when the server stops
func CloseRedis() {
	if Client != nil {
		Client.Close()
		fmt.Println("Redis connection closed.")
	}
}
