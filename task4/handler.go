package main

import (
	"database/sql"
	"log"
	"net/http"
)

func handler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r)
		case http.MethodPost:
			handlePost(w, r, db)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method is not allowed"))
		}
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	data := loadFromCookies(w, r)
	renderForm(w, data)
}

func handlePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := r.ParseForm(); err != nil {
		log.Println("ParseForm error:", err)
		http.Error(w, "From readig error", http.StatusBadRequest)
		return
	}

	data, errors := validate(r)
	if len(errors) > 0 {
		saveErrorsToCookie(w, data, errors)
		http.Redirect(w, r, "form.cgi", http.StatusFound)
		return
	}
	if err := saveToDatabase(db, data); err != nil {
		log.Println("saveToDatabase error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error. Try again later"))
		return
	}

	saveSuccessToCookie(w, data)
	http.Redirect(w, r, "form.cgi", http.StatusFound)
}
