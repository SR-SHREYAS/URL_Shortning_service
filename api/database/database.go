package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background() // create context

func CreateClient(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{ // rdb is industry standard redis client name initialization
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNo,
	})
	return rdb
}
