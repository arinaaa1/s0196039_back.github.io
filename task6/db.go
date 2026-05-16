package main

import (
	"database/sql"
	"fmt"
)

func saveToDatabase(db *sql.DB, data FormData) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("Begin of transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO applications (full_name, phone, email,
		birth_date, gender, biography, contract_accepted)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`)
	if err != nil {
		return 0, fmt.Errorf("Prepare application insert: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(data.Name, data.Phone, data.Email, data.Birthdate,
		data.Gender, data.Bio)
	if err != nil {
		return 0, fmt.Errorf("Execute application insert: %w", err)
	}

	appID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Get last insert ID: %w", err)
	}

	langSTMT, err := tx.Prepare(`
		INSERT INTO application_languages (application_id, language_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("Prepare language insert: %w", err)
	}
	defer langSTMT.Close()

	for _, lang := range data.Languages {
		if _, err := langSTMT.Exec(appID, lang); err != nil {
			return 0, fmt.Errorf("Execute language insert for lang %s: %w", lang, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("Commit transaction: %w", err)
	}

	return appID, nil
}

func saveCredentials(db *sql.DB, applicationID int64, login, passwordHash string) error {
	stmt, err := db.Prepare(`
		INSERT INTO credentials (application_id, login, password_hash)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("Prepare credentials insert: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(applicationID, login, passwordHash); err != nil {
		return fmt.Errorf("Execute credentials insert: %w", err)
	}

	return nil
}

type Credentials struct {
	ApplicationID int64
	PasswordHash  string
}

func findCredentialsByLogin(db *sql.DB, login string) (Credentials, error) {
	var creds Credentials

	stmt, err := db.Prepare(`
		SELECT application_id, password_hash
		FROM credentials
		WHERE login = ?
	`)
	if err != nil {
		return creds, fmt.Errorf("find credentials by login prepare: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(login).Scan(&creds.ApplicationID, &creds.PasswordHash)
	if err == sql.ErrNoRows {
		return creds, fmt.Errorf("Login not found")
	}
	if err != nil {
		return creds, fmt.Errorf("find credentials by login scan: %w", err)
	}
	return creds, nil
}

func getApplicationByID(db *sql.DB, id int64) (FormData, error) {
	var data FormData

	stmt, err := db.Prepare(`
		SELECT full_name, phone, email, birth_date, gender, biography, contract_accepted
		FROM applications
		WHERE id = ?
	`)
	if err != nil {
		return data, fmt.Errorf("get application by ID prepare: %w", err)
	}
	defer stmt.Close()

	var contractAccepted int
	err = stmt.QueryRow(id).Scan(
		&data.Name,
		&data.Phone,
		&data.Email,
		&data.Birthdate,
		&data.Gender,
		&data.Bio,
		&contractAccepted,
	)
	if err == sql.ErrNoRows {
		return data, fmt.Errorf("Form was not found")
	}
	if err != nil {
		return data, fmt.Errorf("get application by ID scan: %w", err)
	}
	data.Contract = contractAccepted == 1

	langStmt, err := db.Prepare(`
		SELECT language_id
		FROM application_languages
		WHERE application_id = ?
	`)
	if err != nil {
		return data, fmt.Errorf("get application by ID langs prepare: %w", err)
	}
	defer langStmt.Close()

	rows, err := langStmt.Query(id)
	if err != nil {
		return data, fmt.Errorf("get application by ID query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var langID string
		if err := rows.Scan(&langID); err != nil {
			return data, fmt.Errorf("get application langs scan: %w", err)
		}
		data.Languages = append(data.Languages, langID)
	}
	return data, nil
}

func updateApplication(db *sql.DB, id int64, data FormData) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("update application begin: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE applications
		SET full_name = ?, phone = ?, email = ?, birth_date = ?,
		gender = ?, biography = ?, contract_accepted = 1
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("update application prepare: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(
		data.Name, data.Phone, data.Email,
		data.Birthdate, data.Gender, data.Bio, id,
	); err != nil {
		return fmt.Errorf("update application exec: %w", err)
	}
	delStmt, err := tx.Prepare(`
		DELETE FROM application_languages WHERE application_id = ?
	`)
	if err != nil {
		return fmt.Errorf("update application delete langs prepare: %w", err)
	}
	defer delStmt.Close()

	if _, err := delStmt.Exec(id); err != nil {
		return fmt.Errorf("update application delete langs exec: %w", err)
	}

	langStmt, err := tx.Prepare(`
		INSERT INTO application_languages (application_id, language_id) VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("update application insert langs prepare: %w", err)
	}
	defer langStmt.Close()

	for _, langID := range data.Languages {
		if _, err := langStmt.Exec(id, langID); err != nil {
			return fmt.Errorf("update application insert lang exec: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update application commit: %w", err)
	}

	return nil
}

type ApplicationRow struct {
	ID        int64
	Name      string
	Phone     string
	Email     string
	Birthdate string
	Gender    string
	Bio       string
	Contract  bool
	Languages []string
}

func getAllApplications(db *sql.DB) ([]ApplicationRow, error) {
	rows, err := db.Query(`
		SELECT id, full_name, phone, email, birth_date, gender, biography, contract_accepted
		FROM applications
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("getAllApplications query: %w", err)
	}
	defer rows.Close()
	var apps []ApplicationRow
	for rows.Next() {
		var app ApplicationRow
		var contract int
		if err := rows.Scan(
			&app.ID, &app.Name, &app.Phone, &app.Email,
			&app.Birthdate, &app.Gender, &app.Bio, &contract,
		); err != nil {
			return nil, fmt.Errorf("getAllApplications scan: %w", err)
		}
		app.Contract = contract == 1
		apps = append(apps, app)
	}

	langStmt, err := db.Prepare(`
		SELECT pl.name
		FROM application_languages al
		JOIN programming_languages pl ON al.language_id = pl.id
		WHERE al.application_id = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("getAllApplications lang prepare: %w", err)
	}
	defer langStmt.Close()
	for i := range apps {
		langRows, err := langStmt.Query(apps[i].ID)
		if err != nil {
			return nil, fmt.Errorf("getAllApplications lang query: %w", err)
		}

		for langRows.Next() {
			var name string
			if err := langRows.Scan(&name); err != nil {
				langRows.Close()
				return nil, fmt.Errorf("getAllApplications lang scan: %w", err)
			}
			apps[i].Languages = append(apps[i].Languages, name)
		}
		langRows.Close()
	}
	return apps, nil
}

func deleteApplication(db *sql.DB, id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("deleteApplicatim begin: %w", err)
	}
	defer tx.Rollback()

	for _, query := range []string{
		`DELETE FROM application_languages WHERE application_id = ?`,
		`DELETE FROM credentials WHERE application_id = ?`,
		`DELETE FROM applications WHERE id = ?`,
	} {
		if _, err := tx.Exec(query, id); err != nil {
			return fmt.Errorf("deleteApplication exec: %w", err)
		}
	}

	return tx.Commit()
}

type LanguageStat struct {
	Name  string
	Count int
}

func getLanguageStats(db *sql.DB) ([]LanguageStat, error) {
	rows, err := db.Query(`
		SELECT pl.name, COUNT(al.application_id) as count
		FROM programming_languages pl
		LEFT JOIN application_languages al ON pl.id = al.language_id
		GROUP BY pl.id, pl.name
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("getLanguageStats query: %w", err)
	}
	defer rows.Close()

	var stats []LanguageStat
	for rows.Next() {
		var s LanguageStat
		if err := rows.Scan(&s.Name, &s.Count); err != nil {
			return nil, fmt.Errorf("getLanguageStats scan: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func getAdminByLogin(db *sql.DB, login string) (string, error) {
	var passwordHash string
	err := db.QueryRow(`
		SELECT password_hash FROM admin_credentials WHERE login = ?
	`, login).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("admin not found")
	}
	if err != nil {
		return "", fmt.Errorf("getAdminByLogin: %w", err)
	}
	return passwordHash, nil
}
