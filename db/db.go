package db

import (
	"context"
	"errors"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/cache"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/models"
)

type Database struct {
	RedisClient *cache.RedisClient
	Mongo       string
}

var (
	ErrNil = errors.New("no matching record found in redis database")
	Ctx    = context.TODO()
)

func InitDatabases(ctx context.Context, redisConf cache.RedisConfig, logger models.Logger) (*Database, error) {

	redisClient, err := cache.NewRedisCache(ctx, redisConf.Host, redisConf.Db, redisConf.Expires, logger)
	if err != nil {
		logger.Error.Printf("Failed to init Databases -> Redis: %v", err)
		return nil, err
	}
	// client := redis.NewClient(&redis.Options{
	// 	Addr:     address,
	// 	Password: "",
	// 	DB:       0,
	// })

	return &Database{
		RedisClient: &redisClient,
		Mongo:       "unimplemented hehe",
	}, nil
}
