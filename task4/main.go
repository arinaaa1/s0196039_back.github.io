package main

import (
	"database/sql"
	"log"
	"net/http/cgi"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "u82356:9869396@tcp(localhost:3306)/u82356")
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	cgi.Serve(handler(db))
}
