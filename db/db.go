package db

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

func GetDB() *sql.DB {
	once.Do(func() {
		var err error
		dsn := "host=localhost port=5432 user=temporal password=temporal dbname=temporal sslmode=disable"
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	})
	return db
}
