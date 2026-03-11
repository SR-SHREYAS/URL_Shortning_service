package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background() // create context

func CreateClient(dbNo int) *redis.Client {
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		if opt, err := redis.ParseURL(redisURL); err == nil {
			opt.DB = dbNo
			return redis.NewClient(opt)
		}
	}

	addr := os.Getenv("REDIS_ADDRESS")
	if addr == "" {
		addr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{ // rdb is industry standard redis client name initialization
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNo,
	})
	return rdb
}
