package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/cgi"
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

	var newCreds map[string]string
	if raw, ok := getCookieValue(r, "new_credentials"); ok {
		newCreds = make(map[string]string)
		decodeFromCookie(raw, &newCreds)
		deleteCookies(w, "new_credentials")
	}

	renderForm(w, data, newCreds)
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
	appID, err := saveToDatabase(db, data)
	if err != nil {
		log.Println("saveToDatabase error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error. Try again later"))
		return
	}

	login, err := generateLogin()
	if err != nil {
		log.Println("generateLogin:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	password, err := generatePassword()
	if err != nil {
		log.Println("generatePassword:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		log.Println("hashPassword", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := saveCredentials(db, appID, login, passwordHash); err != nil {
		log.Panicln("saveCredentials:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	saveSuccessToCookie(w, data)

	if encoded, err := encodeToCookie(map[string]string{
		"login":    login,
		"password": password,
	}); err == nil {
		setSessionCookie(w, "new_credentials", encoded)
	}

	http.Redirect(w, r, "form.cgi", http.StatusFound)
}

func runForm(db *sql.DB) {
	cgi.Serve(handler(db))
}
