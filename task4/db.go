package main

import (
	"database/sql"
	"fmt"
)

func saveToDatabase(db *sql.DB, data FormData) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Begin of transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO applications (full_name, phone, email,
		birth_date, gender, biography, contract_accepted)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`)
	if err != nil {
		return fmt.Errorf("Prepare application insert: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(data.Name, data.Phone, data.Email, data.Birthdate,
		data.Gender, data.Bio)
	if err != nil {
		return fmt.Errorf("Execute application insert: %w", err)
	}

	appID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("Get last insert ID: %w", err)
	}

	langSTMT, err := tx.Prepare(`
		INSERT INTO application_languages (application_id, language_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("Prepare language insert: %w", err)
	}
	defer langSTMT.Close()

	for _, lang := range data.Languages {
		if _, err := langSTMT.Exec(appID, lang); err != nil {
			return fmt.Errorf("Execute language insert for lang %s: %w", lang, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Commit transaction: %w", err)
	}

	return nil
}
