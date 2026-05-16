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
		log.Fatal("Database is missing")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	switch filepath.Base(os.Args[0]) {
	case "form.cgi":
		runForm(db)
	case "login.cgi":
		runLogin(db)
	case "edit.cgi":
		runEdit(db)
	case "logout.cgi":
		runLogout()
	case "admin.cgi":
		runAdmin(db)
	default:
		runIndex()
	}
}
