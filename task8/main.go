package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN не найден")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("sql.Open:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("db.Ping:", err)
	}
	switch filepath.Base(os.Args[0]) {
	case "api.cgi":
		runAPI(db)
	default:
		log.Fatal("неизвестный бинарный файл")
	}
}
