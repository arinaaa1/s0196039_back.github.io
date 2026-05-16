package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"regexp"
	"strings"
)

type APIRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Consent bool   `json:"consent"`
}

type APIUpdateRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Login      string `json:"login,omitempty"`
	Password   string `json:"password,omitempty"`
	Token      string `json:"token,omitempty"`
	ProfileURL string `json:"profile_url,omitempty"`
}

var (
	reAPIName  = regexp.MustCompile(`^[\p{L} ]+$`)
	reAPIPhone = regexp.MustCompile(`^\+?[0-9()\- ]{7,32}$`)
	reAPIEmail = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-type, Authorization")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func validateAPIRequest(req APIRequest) string {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return "Имя обязательно для заполнения"
	}
	if !reAPIName.MatchString(name) {
		return "Имя может содержать только буквы и пробелы"
	}

	phone := strings.TrimSpace(req.Phone)
	if phone != "" && !reAPIPhone.MatchString(phone) {
		return "Некорректный телефон"
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		return "Email обязателен для заполнения"
	}
	if !reAPIEmail.MatchString(email) {
		return "Email указан некорректно"
	}

	if !req.Consent {
		return "Необходимо согласние на обработку персональных данных"
	}
	return ""
}

func validateAPIUpdateRequest(req APIUpdateRequest) string {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return "Имя обязательно для заполнения"
	}
	if !reAPIName.MatchString(name) {
		return "Имя может содержать только буквы и пробелы"
	}

	phone := strings.TrimSpace(req.Phone)
	if phone != "" && !reAPIPhone.MatchString(phone) {
		return "Некорректный телефон"
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		return "Email обязателен для заполнения"
	}
	if !reAPIEmail.MatchString(email) {
		return "Email указан некорректно"
	}
	return ""
}

func runAPI(db *sql.DB) {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		action := r.URL.Query().Get("action")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		switch {
		case r.Method == http.MethodGet && action == "profile":
			handleAPIProfile(w, r, db)
		case r.Method == http.MethodPost && action == "login":
			handleAPILogin(w, r, db)
		case r.Method == http.MethodPost:
			handleAPIPost(w, r, db)
		case r.Method == http.MethodPut:
			handleAPIPut(w, r, db)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
				Error: "Метод не поддерживается",
			})
		}
	}))
}

func handleAPIProfile(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = os.Getenv("HTTP_AUTHORIZATION")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, APIResponse{
			Error: "Требуется авторизация",
		})
		return
	}
	payload, err := validateJWT(authHeader[7:])
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{
			Error: "Недействительный токен",
		})
		return
	}

	app, err := getAPIApplicationByID(db, payload.ApplicationID)
	if err != nil {
		log.Println("getAPIApplicationByID:", err)
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Error: "Внутренняя ошибка сервера",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"name":    app.Name,
		"phone":   app.Phone,
		"email":   app.Email,
		"message": app.Message,
	})
}

func handleAPILogin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Error: "Неверный формат JSON",
		})
		return
	}

	creds, err := findAPICredentialsByLogin(db, req.Login)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{
			Error: "Неверный логин или пароль",
		})
		return
	}

	if !checkPassword(req.Password, creds.Passwordhash) {
		writeJSON(w, http.StatusUnauthorized, APIResponse{
			Error: "Неверный логин или пароль",
		})
		return
	}

	token, err := generateJWT(creds.ApplicationID, creds.Login)
	if err != nil {
		log.Println("generateJWT:", err)
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Error: "Внутренняя ошибка сервера",
		})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Login:   creds.Login,
		Token:   token,
	})
}

func handleAPIPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req APIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Error: "Неверный формат JSON",
		})
		return
	}

	if errMsg := validateAPIRequest(req); errMsg != "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Error: errMsg,
		})
		return
	}

	appID, err := saveAPIApplication(db, req)
	if err != nil {
		log.Println("saveAPIApplication:", err)
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Error: "Внутренняя ошибка сервера",
		})
		return
	}

	login, err := generateLogin()
	if err != nil {
		log.Println("generateLogin:", err)
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Error: "Внутренняя ошибка сервера",
		})
		return
	}

	password, err := generatePassword()
	if err != nil {
		log.Println("generatePassword:", err)
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Error: "Внутренняя ошибка сервера",
		})
		return
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		log.Println("hashPassword:", err)
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Error: "Внутренняя ошибка сервера",
		})
		return
	}

	if err := saveAPICredentials(db, appID, login, passwordHash); err != nil {
		log.Println("saveAPICredentials:", err)
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Error: "Внутренняя ошибка сервера",
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success:    true,
		Login:      login,
		Password:   password,
		ProfileURL: "/web_backend/task8/edit.cgi",
	})
}

func handleAPIPut(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = os.Getenv("HTTP_AUTHORIZATION")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, APIResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	tokenStr := authHeader[7:]
	payload, err := validateJWT(tokenStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{
			Error: "Недействительный токен",
		})
		return
	}

	var req APIUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Error: "Неверный формат JSON",
		})
		return
	}

	if errMsg := validateAPIUpdateRequest(req); errMsg != "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Error: errMsg,
		})
		return
	}
	if err := updateApiApplication(db, payload.ApplicationID, req); err != nil {
		log.Println("updateApiApplication:", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
	})
}
