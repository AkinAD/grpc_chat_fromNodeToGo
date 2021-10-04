package database

import (
	"errors"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/config"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/database/cache"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/database/db"
	manager "github.com/AkinAD/grpc_chat_fromNodeToGo/models/Manager"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	RedisClient cache.RedisCache
	SQLite       db.SQLiteDB
}

var (
	ErrNil = errors.New("no matching record found in redis database")
)

func InitDatabases(m manager.Manager, redisConf config.RedisConfig, liteConfig config.SQLiteConfig, ) (*Database, error) {
	logger := m.Logger
	logger.Info.Print("Initializing database clients...")
	redisClient, err := initRedis(m, redisConf)
	if err != nil {
		return nil, err
	}
	logger.Info.Print("Successfully created Redis client.")

	db, err := initSqLite(m, liteConfig )
	if err != nil {
		return nil, err
	}
	logger.Info.Print("Successfully created SQLite client.")

	return &Database{
		RedisClient: *redisClient,
		SQLite:       db,
	}, nil
}


func initRedis(m manager.Manager, redisConf config.RedisConfig) (*cache.RedisCache, error) {
	logger := m.Logger
	redisClient, err := cache.NewRedisCache(&m, redisConf.Host, redisConf.Db, redisConf.Expires)
	if err != nil {
		logger.Error.Printf("Failed to init Databases -> Redis: %v", err)
		return nil, err
	}
	// client := redis.NewClient(&redis.Options{
	// 	Addr:     address,
	// 	Password: "",
	// 	DB:       0,
	// })
	return &redisClient, nil
}

//TODO: change to some other db later
func initSqLite( m manager.Manager, liteConfig config.SQLiteConfig, ) (db.SQLiteDB, error) {
	logger := m.Logger
	sqLiteClient, err := db.NewSQLiteDB(liteConfig, m)
	if err != nil {
		logger.Error.Printf("Failed to init Databases -> SQLite: %v", err)
		return nil, err
	}
	return sqLiteClient, nil
}