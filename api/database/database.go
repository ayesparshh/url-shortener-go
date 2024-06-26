package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

func CreateClient(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{    
		Addr:     os.Getenv("Dbaddress"),
		Password: os.Getenv("Dbpassword"),
		DB : dbNo,
	})

	return rdb
}

