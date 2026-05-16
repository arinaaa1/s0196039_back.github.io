package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/http/cgi"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/go-sql-driver/mysql"
)

type FormData struct {
	Name      string
	Phone     string
	Email     string
	Birthdate string
	Gender    string
	Bio       string
	Languages []string
}

var validGenders = map[string]bool{
	"male": true, "female": true,
}

var validLanguageIDs = map[string]bool{
	"1": true, "2": true, "3": true, "4": true, "5": true,
	"6": true, "7": true, "8": true, "9": true, "10": true,
	"11": true, "12": true,
}

func validate(r *http.Request) (FormData, []string) {
	var data FormData
	var errors []string

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		errors = append(errors, "Name is required")
	} else if utf8.RuneCountInString(name) > 150 {
		errors = append(errors, "Name must be at most 150 characters")
	} else if ok, _ := regexp.MatchString(`^[\p{L} ]+$`, name); !ok {
		errors = append(errors, "Name contains invalid characters")
	} else {
		data.Name = name
	}

	phone := strings.TrimSpace(r.FormValue("phone"))
	if phone == "" {
		errors = append(errors, "Phone is required")
	} else if ok, _ := regexp.MatchString(`^\+?[0-9()\- ]{7,32}$`, phone); !ok {
		errors = append(errors, "Phone format is invalid")
	} else {
		data.Phone = phone
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		errors = append(errors, "Email is required")
	} else if ok, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email); !ok {
		errors = append(errors, "Email format is invalid")
	} else {
		data.Email = email
	}

	birthdate := strings.TrimSpace(r.FormValue("birthdate"))
	if birthdate == "" {
		errors = append(errors, "Birthdate is required")
	} else if _, err := time.Parse("2006-01-02", birthdate); err != nil {
		errors = append(errors, "Birthdate format is invalid (expected YYYY-MM-DD)")
	} else {
		data.Birthdate = birthdate
	}
	gender := strings.TrimSpace(r.FormValue("gender"))
	if !validGenders[gender] {
		errors = append(errors, "Gender must be 'male' or 'female'")
	} else {
		data.Gender = gender
	}

	languages := r.Form["languages[]"]
	if len(languages) == 0 {
		errors = append(errors, "At least one language must be selected")
	} else {
		allValid := true
		for _, lang := range languages {
			if !validLanguageIDs[lang] {
				errors = append(errors, "Invalid language selection"+html.EscapeString(lang))
				allValid = false
				break
			}
		}
		if allValid {
			data.Languages = languages
		}
	}

	bio := strings.TrimSpace(r.FormValue("bio"))
	if bio == "" {
		errors = append(errors, "Bio is required")
	} else {
		data.Bio = bio
	}

	if r.FormValue("contract") == "" {
		errors = append(errors, "You must agree to the contract")
	}

	return data, errors
}

func saveToDatabase(db *sql.DB, data FormData) error {
	stmt, err := db.Prepare(`
		INSERT INTO applications (full_name, phone, email,
		birth_date, gender, biography, contract_accepted)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(data.Name, data.Phone, data.Email, data.Birthdate, data.Gender, data.Bio)
	if err != nil {
		return err
	}

	appID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	langSTMT, err := db.Prepare(`
		INSERT INTO application_languages (application_id, language_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer langSTMT.Close()

	for _, lang := range data.Languages {
		if _, err := langSTMT.Exec(appID, lang); err != nil {
			return err
		}
	}

	return nil
}

func makeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		data, errors := validate(r)
		if len(errors) > 0 {
			fmt.Fprintln(w, "<h2>Erorrs:</h2><ul>")
			for _, err := range errors {
				fmt.Fprintf(w, "<li>%s</li>", err)
			}
			fmt.Fprintln(w, "</ul>")
			return
		}

		if err := saveToDatabase(db, data); err != nil {
			fmt.Fprintln(w, "<h2>Database error:</h2><p>"+html.EscapeString(err.Error())+"</p>")
			return
		}

		fmt.Fprintln(w, "<h2>Application submitted successfully!</h2>")
	}
}

func main() {
	db, err := sql.Open("mysql", "u82944:dw4$Ewe2@tcp(localhost:3306)/u82944")
	if err != nil {
		log.Fatal("Database connection error: ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Database ping error: ", err)
	}

	cgi.Serve(makeHandler(db))
}
