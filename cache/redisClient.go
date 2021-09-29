package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/models"
	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Host    string
	Db      int
	Expires time.Duration
}

type RedisCache struct {
	redisConfig RedisConfig
	ctx         context.Context
	Client      redis.Client
	Logger      models.Logger
}

type RedisClient interface {
	Set(key string, value *[]byte)
	Get(key string) *[]byte
}

func NewRedisCache(ctx context.Context, host string, db int, exp time.Duration, logger models.Logger) (RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error.Printf("Error reaching redis: %v", err)
		return nil, err
	}

	return &RedisCache{
		redisConfig: RedisConfig{Host: host, Db: db, Expires: exp},
		ctx:         ctx,
		Client:      *client,
	}, nil
}

func (cache *RedisCache) Set(key string, value *[]byte) {
	client := cache.Client

	// if err != nil {
	// 	log.Errorf("Unable to marshal data for cahce set: %v", err)
	// }
	client.Set(cache.ctx, key, value, cache.redisConfig.Expires)

}

func (cache *RedisCache) Get(key string) *[]byte {
	client := cache.Client

	val, err := client.Get(cache.ctx, key).Result()
	if err != nil {
		cache.Logger.Error.Printf("Error fetching from cache: %v", err)
		return nil
	}
	res := []byte(val)
	return &res
}

func (cache *RedisCache) GetSession(key string) *models.Session {
	session := models.Session{}
	fetched := cache.Get(key)
	err := json.Unmarshal(*fetched, &session)
	if err != nil {
		cache.Logger.Error.Panicf("Unable to unmarshal value: %v", err)
		return nil
	}

	return &session
}

func (cache *RedisCache) SetSession(key string, sesh models.Session) {
	json, err := json.Marshal(sesh)
	if err != nil {
		cache.Logger.Error.Panicf("Unable to Marshal value: %v", err)
	}
	cache.Set(key, &json)

}
