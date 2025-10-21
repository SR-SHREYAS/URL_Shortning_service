package database

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var (
	// Ctx is a package-level context, though using request-specific context is often better.
	Ctx = context.Background()

	// Rdb0 and Rdb1 are exported (capitalized) and will hold our shared client instances.
	Rdb0 *redis.Client
	Rdb1 *redis.Client
)

// createClient is an unexported helper function to create a new Redis client.
func createClient(dbNo int) *redis.Client {
	// Use redis.ParseURL to correctly handle full connection strings from services like Render.
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		panic(fmt.Sprintf("Failed to parse Redis URL: %v", err))
	}

	// Set the specific database number for this client instance.
	opt.DB = dbNo

	rdb := redis.NewClient(opt)
	// Ping the client to ensure the connection is valid.
	if err := rdb.Ping(Ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis DB %d: %v", dbNo, err))
	}

	return rdb
}

// Init initializes the database clients. It should be called once at startup.
func Init() {
	Rdb0 = createClient(0) // For URLs
	Rdb1 = createClient(1) // For rate limiting and counters
}
