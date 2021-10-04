package config

import "time"

type RedisConfig struct {
	Host    string
	Db      int
	Expires time.Duration
}

type SQLiteConfig struct {
	DriverName     string
	DataSourceName string
}