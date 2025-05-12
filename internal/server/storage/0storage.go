package storage

import (
	"database/sql"
	"log"

	"github.com/kw3a/spotted-server/internal/database"
)

type MysqlStorage struct {
	Queries *database.Queries
	db      *sql.DB
}

func NewMysqlStorage(dbURL string) (*MysqlStorage, error) {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	queries := database.New(db)

	return &MysqlStorage{Queries: queries, db: db}, nil
}
