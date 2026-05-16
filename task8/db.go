package main

import (
	"database/sql"
	"fmt"
)

func saveAPIApplication(db *sql.DB, req APIRequest) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO applications_v2 (name, phone, email, message, contract_accepted)
		VALUES (?,?,?,?,1)
	`)
	if err != nil {
		return 0, fmt.Errorf("saveAPIApplication prepare: %w", err)
	}
	defer stmt.Close()
	result, err := stmt.Exec(
		req.Name,
		req.Phone,
		req.Email,
		req.Message,
	)

	if err != nil {
		return 0, fmt.Errorf("saveAPIApplication exec: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("saveAPIApplication lastInsertID: %w", err)
	}

	return id, nil
}

func saveAPICredentials(db *sql.DB, applicationID int64, login, passwordHash string) error {
	stmt, err := db.Prepare(`
		INSERT INTO credentials_v2 (application_id, login, password_hash)
		VALUES (?, ?, ?)
	`)

	if err != nil {
		return fmt.Errorf("saveAPICredentials prepare: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(applicationID, login, passwordHash); err != nil {
		return fmt.Errorf("saveAPICredentials exec: %w", err)
	}
	return nil
}

type APICredentials struct {
	ApplicationID int64
	Passwordhash  string
	Login         string
}

func findAPICredentialsByLogin(db *sql.DB, login string) (APICredentials, error) {
	var creds APICredentials
	err := db.QueryRow(`
		SELECT application_id, password_hash, login
		FROM credentials_v2
		WHERE login = ?
	`, login).Scan(&creds.ApplicationID, &creds.Passwordhash, &creds.Login)

	if err == sql.ErrNoRows {
		return creds, fmt.Errorf("Логин не найден")
	}
	if err != nil {
		return creds, fmt.Errorf("findAPICredentialsByLogin: %w", err)
	}

	return creds, nil
}

func getAPIApplicationByID(db *sql.DB, id int64) (APIUpdateRequest, error) {
	var req APIUpdateRequest

	err := db.QueryRow(`
		SELECT name, phone, email, message
		FROM applications_v2
		WHERE id = ?
	`, id).Scan(&req.Name, &req.Phone, &req.Email, &req.Message)

	if err == sql.ErrNoRows {
		return req, fmt.Errorf("Заявка не найдена")
	}
	if err != nil {
		return req, fmt.Errorf("getAPIApplicationByID: %w", err)
	}

	return req, nil
}

func updateApiApplication(db *sql.DB, id int64, req APIUpdateRequest) error {
	stmt, err := db.Prepare(`
		UPDATE applications_v2
		SET name = ?, phone = ?, email = ?, message = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("updateAPIApplication prepare: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(req.Name, req.Phone, req.Email, req.Message, id); err != nil {
		return fmt.Errorf("updateAPIApplication exec: %w", err)
	}
	return nil
}
