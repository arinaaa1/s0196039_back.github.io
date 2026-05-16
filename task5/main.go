package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "u82944:dw4$Ewe2@tcp(localhost:3306)/u82944")
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
	default:
		runIndex()
	}
}
