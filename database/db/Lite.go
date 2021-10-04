package db

import (
	"database/sql"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/config"
	manager "github.com/AkinAD/grpc_chat_fromNodeToGo/models/Manager"
) 


type SQLiteClient struct {
	Client       *sql.DB
	Manager manager.Manager
}


type SQLiteDB interface {
	RoomRepository() *RoomRepository
	UserRepository() *UserRepository
}


func NewSQLiteDB(cfg config.SQLiteConfig, m manager.Manager) (SQLiteDB, error) {
	logger := m.Logger

	db, err := sql.Open(cfg.DriverName, cfg.DataSourceName)
	if err != nil {
		logger.Error.Fatal(err)
		return nil, err
	}

	sqlStmt := `	
	CREATE TABLE IF NOT EXISTS room (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		private TINYINT NULL
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		logger.Error.Fatalf("%q: %s\n", err, sqlStmt)
		return nil, err

	}

	sqlStmt = `	
	CREATE TABLE IF NOT EXISTS user (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		username VARCHAR(255)  NULL,
		password VARCHARR(255)  NULL
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		logger.Error.Fatalf("%q: %s\n", err, sqlStmt)
		return nil, err

	}
	return &SQLiteClient{Client: db, Manager: m}, nil
}


func (s SQLiteClient) RoomRepository() *RoomRepository {
	return &RoomRepository{Db: s.Client}
}

func (s SQLiteClient) UserRepository() *UserRepository {
	return &UserRepository{Db: s.Client}
}