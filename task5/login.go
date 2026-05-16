package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"net/http/cgi"
)

type LoginPageData struct {
	Error string
}

var loginTemplate = template.Must(template.New("login").Parse(loginHTML))

func renderLogin(w http.ResponseWriter, data LoginPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	loginTemplate.Execute(w, data)
}

func runLogin(db *sql.DB) {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLoginGet(w, r)
		case http.MethodPost:
			handleLoginPost(w, r, db)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func handleLoginGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := getJWTFromCookie(r); ok {
		http.Redirect(w, r, "edit.cgi", http.StatusFound)
		return
	}
	renderLogin(w, LoginPageData{})
}

func handleLoginPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")

	if login == "" || password == "" {
		renderLogin(w, LoginPageData{Error: "Login and password cannot be empty"})
		return
	}

	creds, err := findCredentialsByLogin(db, login)
	if err != nil {
		log.Println("findCredentialsByLogin:", err)
		renderLogin(w, LoginPageData{Error: "Invalid login or password"})
		return
	}

	if !checkPassword(password, creds.PasswordHash) {
		renderLogin(w, LoginPageData{Error: "Invalid login or password"})
		return
	}

	token, err := generateJWT(creds.ApplicationID, login)
	if err != nil {
		log.Println("generateJWT:", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	setJWTCookie(w, token)
	http.Redirect(w, r, "edit.cgi", http.StatusFound)
}

const loginHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Вход</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }

		body {
			font-family: Arial, sans-serif;
			background: #f0f2f5;
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
		}

		.card {
			background: white;
			border-radius: 12px;
			box-shadow: 0 4px 24px rgba(0,0,0,0.10);
			padding: 40px;
			width: 100%;
			max-width: 400px;
		}

		.logo {
			text-align: center;
			margin-bottom: 28px;
		}

		.logo-icon {
			font-size: 48px;
			display: block;
			margin-bottom: 8px;
		}

		h1 {
			font-size: 22px;
			font-weight: 700;
			color: #1a1a2e;
			text-align: center;
			margin-bottom: 24px;
		}

		.field {
			margin-bottom: 18px;
		}

		label {
			display: block;
			font-size: 13px;
			font-weight: 600;
			color: #555;
			margin-bottom: 6px;
			text-transform: uppercase;
			letter-spacing: 0.4px;
		}

		input {
			width: 100%;
			padding: 10px 14px;
			border: 1.5px solid #d1d5db;
			border-radius: 8px;
			font-size: 15px;
			color: #222;
			background: #fafafa;
			outline: none;
			transition: border-color 0.2s;
			font-family: inherit;
		}

		input:focus {
			border-color: #2c7be5;
			background: #fff;
		}

		.error-banner {
			background: #fff5f5;
			border: 1.5px solid #e53e3e;
			border-radius: 8px;
			padding: 12px 16px;
			color: #e53e3e;
			font-size: 14px;
			margin-bottom: 20px;
			text-align: center;
		}

		.btn {
			width: 100%;
			padding: 12px;
			background: #2c7be5;
			color: white;
			border: none;
			border-radius: 8px;
			font-size: 16px;
			font-weight: 600;
			cursor: pointer;
			transition: background 0.2s;
			font-family: inherit;
		}

		.btn:hover { background: #1a5fc1; }

		.links {
			text-align: center;
			margin-top: 20px;
			font-size: 14px;
			color: #666;
		}

		.links a {
			color: #2c7be5;
			text-decoration: none;
			font-weight: 500;
		}

		.links a:hover { text-decoration: underline; }
	</style>
</head>
<body>
<div class="card">
	<div class="logo">
		<span class="logo-icon">🔐</span>
	</div>
	<h1>Вход в систему</h1>

	{{if .Error}}
	<div class="error-banner">{{.Error}}</div>
	{{end}}

	<form action="login.cgi" method="POST">
		<div class="field">
			<label>Логин</label>
			<input type="text" name="login" autocomplete="username">
		</div>
		<div class="field">
			<label>Пароль</label>
			<input type="password" name="password" autocomplete="current-password">
		</div>
		<button type="submit" class="btn">Войти</button>
	</form>

	<div class="links">
		<a href="form.cgi">← Заполнить новую анкету</a>
	</div>
</div>
</body>
</html>`
