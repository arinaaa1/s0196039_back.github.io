package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"net/http/cgi"
)

type EditPageData struct {
	PageData
	Login string
}

var editTemplate = template.Must(template.New("edit").Parse(editHTML))

func renderEdit(w http.ResponseWriter, data EditPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data.Languages = allLanguages
	editTemplate.Execute(w, data)
}

func requireAuth(w http.ResponseWriter, r *http.Request) (JWTPayload, bool) {
	payload, ok := getJWTFromCookie(r)
	if !ok {
		http.Redirect(w, r, "login.cgi", http.StatusFound)
		return JWTPayload{}, false
	}
	return payload, true
}

func runEdit(db *sql.DB) {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleEditGet(w, r, db)
		case http.MethodPost:
			handleEditPost(w, r, db)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func handleEditGet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	payload, ok := requireAuth(w, r)
	if !ok {
		return
	}

	pageData := loadFromCookies(w, r)

	if len(pageData.Errors) == 0 {
		appData, err := getApplicationByID(db, payload.ApplicationID)
		if err != nil {
			log.Println("getApplicarionByID:", err)
			http.Error(w, "Failed to load application data", http.StatusInternalServerError)
			return
		}
		pageData.Values = appData
	}

	renderEdit(w, EditPageData{
		PageData: pageData,
		Login:    payload.Login,
	})
}

func handleEditPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	payload, ok := requireAuth(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	data, errors := validate(r)

	if len(errors) > 0 {
		saveErrorsToCookie(w, data, errors)
		http.Redirect(w, r, "edit.cgi", http.StatusNotFound)
		return
	}

	if err := updateApplication(db, payload.ApplicationID, data); err != nil {
		log.Println("updateApplication:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	saveSuccessToCookie(w, data)
	http.Redirect(w, r, "edit.cgi", http.StatusFound)
}

const editHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Редактирование анкеты</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }

		body {
			font-family: Arial, sans-serif;
			background: #f0f2f5;
			min-height: 100vh;
			padding: 40px 20px;
		}

		.topbar {
			max-width: 560px;
			margin: 0 auto 20px;
			display: flex;
			justify-content: space-between;
			align-items: center;
		}

		.topbar-user {
			font-size: 14px;
			color: #555;
		}

		.topbar-user strong {
			color: #1a1a2e;
		}

		.topbar a {
			font-size: 14px;
			color: #e53e3e;
			text-decoration: none;
			font-weight: 500;
		}

		.topbar a:hover { text-decoration: underline; }

		.card {
			background: white;
			border-radius: 12px;
			box-shadow: 0 4px 24px rgba(0,0,0,0.10);
			padding: 40px;
			width: 100%;
			max-width: 560px;
			margin: 0 auto;
		}

		h1 {
			font-size: 24px;
			font-weight: 700;
			color: #1a1a2e;
			margin-bottom: 28px;
			text-align: center;
		}

		.field { margin-bottom: 20px; }

		.field > label {
			display: block;
			font-size: 13px;
			font-weight: 600;
			color: #555;
			margin-bottom: 6px;
			text-transform: uppercase;
			letter-spacing: 0.4px;
		}

		input[type="text"],
		input[type="tel"],
		input[type="email"],
		input[type="date"],
		select,
		textarea {
			width: 100%;
			padding: 10px 14px;
			border: 1.5px solid #d1d5db;
			border-radius: 8px;
			font-size: 15px;
			color: #222;
			background: #fafafa;
			transition: border-color 0.2s, background 0.2s;
			outline: none;
			font-family: inherit;
		}

		input:focus, select:focus, textarea:focus {
			border-color: #2c7be5;
			background: #fff;
		}

		.field-error input,
		.field-error select,
		.field-error textarea {
			border-color: #e53e3e;
			background: #fff5f5;
		}

		.error-msg {
			font-size: 12px;
			color: #e53e3e;
			margin-top: 5px;
			display: flex;
			align-items: center;
			gap: 4px;
		}

		.error-msg::before {
			content: "!";
			display: inline-flex;
			align-items: center;
			justify-content: center;
			width: 14px;
			height: 14px;
			background: #e53e3e;
			color: white;
			border-radius: 50%;
			font-size: 10px;
			font-weight: 700;
			flex-shrink: 0;
		}

		textarea { height: 110px; resize: vertical; }
		select[multiple] { height: 160px; padding: 6px; }

		select[multiple] option {
			padding: 4px 8px;
			border-radius: 4px;
		}

		select[multiple] option:checked {
			background: #2c7be5;
			color: white;
		}

		.radio-group {
			display: flex;
			gap: 16px;
			margin-top: 2px;
		}

		.radio-group label,
		.checkbox-label {
			display: flex;
			align-items: center;
			gap: 7px;
			font-size: 15px;
			font-weight: 400;
			color: #333;
			cursor: pointer;
			text-transform: none;
			letter-spacing: 0;
		}

		input[type="radio"],
		input[type="checkbox"] {
			width: 16px;
			height: 16px;
			accent-color: #2c7be5;
			cursor: pointer;
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
			margin-top: 8px;
			transition: background 0.2s;
			font-family: inherit;
		}

		.btn:hover { background: #1a5fc1; }

		.success-banner {
			background: #f0fff4;
			border: 1.5px solid #38a169;
			border-radius: 8px;
			padding: 14px 18px;
			color: #276749;
			font-size: 15px;
			margin-bottom: 24px;
			text-align: center;
			font-weight: 500;
		}
	</style>
</head>
<body>

<div class="topbar">
	<span class="topbar-user">Вы вошли как <strong>{{.Login}}</strong></span>
	<a href="logout.cgi">Выйти</a>
</div>

<div class="card">
	<h1>✏️ Редактирование анкеты</h1>

	{{if .Success}}
	<div class="success-banner">✅ Данные успешно обновлены!</div>
	{{end}}

	<form action="edit.cgi" method="POST">

		<div class="field {{if index .Errors "name"}}field-error{{end}}">
			<label>ФИО</label>
			<input type="text" name="name" value="{{.Values.Name}}">
			{{if index .Errors "name"}}
				<div class="error-msg">{{index .Errors "name"}}</div>
			{{end}}
		</div>

		<div class="field {{if index .Errors "phone"}}field-error{{end}}">
			<label>Телефон</label>
			<input type="tel" name="phone" value="{{.Values.Phone}}">
			{{if index .Errors "phone"}}
				<div class="error-msg">{{index .Errors "phone"}}</div>
			{{end}}
		</div>

		<div class="field {{if index .Errors "email"}}field-error{{end}}">
			<label>Email</label>
			<input type="email" name="email" value="{{.Values.Email}}">
			{{if index .Errors "email"}}
				<div class="error-msg">{{index .Errors "email"}}</div>
			{{end}}
		</div>

		<div class="field {{if index .Errors "birthdate"}}field-error{{end}}">
			<label>Дата рождения</label>
			<input type="date" name="birthdate" value="{{.Values.Birthdate}}">
			{{if index .Errors "birthdate"}}
				<div class="error-msg">{{index .Errors "birthdate"}}</div>
			{{end}}
		</div>

		<div class="field {{if index .Errors "gender"}}field-error{{end}}">
			<label>Пол</label>
			<div class="radio-group">
				<label>
					<input type="radio" name="gender" value="male"
						{{if eq .Values.Gender "male"}}checked{{end}}> Мужской
				</label>
				<label>
					<input type="radio" name="gender" value="female"
						{{if eq .Values.Gender "female"}}checked{{end}}> Женский
				</label>
			</div>
			{{if index .Errors "gender"}}
				<div class="error-msg">{{index .Errors "gender"}}</div>
			{{end}}
		</div>

		<div class="field {{if index .Errors "languages"}}field-error{{end}}">
			<label>Любимый язык программирования</label>
			<select name="languages[]" multiple>
				{{range .Languages}}
				<option value="{{.ID}}"
					{{if $.IsSelectedLang .ID}}selected{{end}}>{{.Name}}</option>
				{{end}}
			</select>
			{{if index .Errors "languages"}}
				<div class="error-msg">{{index .Errors "languages"}}</div>
			{{end}}
		</div>

		<div class="field {{if index .Errors "bio"}}field-error{{end}}">
			<label>Биография</label>
			<textarea name="bio">{{.Values.Bio}}</textarea>
			{{if index .Errors "bio"}}
				<div class="error-msg">{{index .Errors "bio"}}</div>
			{{end}}
		</div>

		<div class="field {{if index .Errors "contract"}}field-error{{end}}">
			<label class="checkbox-label">
				<input type="checkbox" name="contract"
					{{if .Values.Contract}}checked{{end}}> С контрактом ознакомлен(а)
			</label>
			{{if index .Errors "contract"}}
				<div class="error-msg">{{index .Errors "contract"}}</div>
			{{end}}
		</div>

		<button type="submit" class="btn">Сохранить изменения</button>
	</form>
</div>
</body>
</html>`
