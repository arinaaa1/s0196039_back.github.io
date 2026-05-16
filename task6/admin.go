package main

import (
	"database/sql"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"strconv"
	"strings"
)

type AdminPageData struct {
	Applications []ApplicationRow
	Stats        []LanguageStat
	EditApp      *ApplicationRow
	EditData     FormData
	EditErrors   FormErrors
	EditID       int64
	CSRFToken    string
}

var adminTmpl = template.Must(template.New("admin").Funcs(template.FuncMap{
	"join": strings.Join,
	"mul":  func(a, b int) int { return a * b },
	"div": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	},
}).Parse(adminHTML))

func requireBasicAuth(w http.ResponseWriter, r *http.Request, db *sql.DB) bool {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		authHeader = os.Getenv("HTTP_AUTHORIZATION")
	}

	if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
		sendUnauthorized(w)
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(authHeader[6:])
	if err != nil {
		sendUnauthorized(w)
		return false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		sendUnauthorized(w)
		return false
	}

	login, password := parts[0], parts[1]

	passwordHash, err := getAdminByLogin(db, login)
	if err != nil {
		sendUnauthorized(w)
		return false
	}
	if !checkPassword(password, passwordHash) {
		sendUnauthorized(w)
		return false
	}

	return true
}

func sendUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Admin Panel"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Need authentication"))
}

func runAdmin(db *sql.DB) {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireBasicAuth(w, r, db) {
			return
		}
		action := r.URL.Query().Get("action")

		switch {
		case r.Method == http.MethodGet && action == "edit":
			handleAdminEditGet(w, r, db)
		case r.Method == http.MethodPost && action == "edit":
			handleAdminEditPost(w, r, db)
		case r.Method == http.MethodPost && action == "delete":
			handleAdminDelete(w, r, db)
		default:
			handleAdminList(w, r, db)
		}
	}))
}

func handleAdminList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	apps, err := getAllApplications(db)
	if err != nil {
		log.Println("getAllApplications: ", err)
		http.Error(w, "Data receiving error", http.StatusInternalServerError)
		return
	}

	stats, err := getLanguageStats(db)
	if err != nil {
		log.Println("getLanguageStats:", err)
		http.Error(w, "Error receiving stats", http.StatusInternalServerError)
		return
	}

	renderAdmin(w, AdminPageData{
		Applications: apps,
		Stats:        stats,
		CSRFToken:    getOrCreateCSRFToken(w, r),
	})
}

func handleAdminEditGet(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	app, err := getApplicationByID(db, id)
	if err != nil {
		log.Println("getApplicationByID:", err)
		http.Error(w, "form was not found", http.StatusNotFound)
		return
	}
	renderAdmin(w, AdminPageData{
		EditData:  app,
		EditID:    id,
		CSRFToken: getOrCreateCSRFToken(w, r),
	})
}

func handleAdminEditPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Reading form error", http.StatusBadRequest)
		return
	}

	if !validateCSRFToken(r) {
		http.Error(w, "Request denied", http.StatusForbidden)
		return
	}

	data, errors := validate(r)

	if len(errors) > 0 {
		renderAdmin(w, AdminPageData{
			EditData:   data,
			EditErrors: errors,
			EditID:     id,
		})
		return
	}

	if err := updateApplication(db, id, data); err != nil {
		log.Println("updateApplication:", err)
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "admin.cgi", http.StatusFound)
}

func handleAdminDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Reading form error", http.StatusBadRequest)
		return
	}

	if !validateCSRFToken(r) {
		http.Error(w, "Request denied", http.StatusForbidden)
		return
	}

	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := deleteApplication(db, id); err != nil {
		log.Println("deleteApplication:", err)
		http.Error(w, "Delete error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "admin.cgi", http.StatusFound)
}

func parseID(r *http.Request) (int64, error) {
	idStr := r.URL.Query().Get("id")
	return strconv.ParseInt(idStr, 10, 64)
}

func renderAdmin(w http.ResponseWriter, data AdminPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminTmpl.Execute(w, struct {
		AdminPageData
		Languages []Language
	}{data, allLanguages}); err != nil {
		log.Println("adminTmpl.Execute:", err)
	}
}

func (d AdminPageData) IsSelectedEditLang(id string) bool {
	for _, selected := range d.EditData.Languages {
		if selected == id {
			return true
		}
	}
	return false
}

const adminHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Админ панель</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }

		body {
			font-family: Arial, sans-serif;
			background: #f0f2f5;
			padding: 30px 20px;
		}

		h1, h2 {
			color: #1a1a2e;
			margin-bottom: 20px;
		}

		h1 { font-size: 26px; }
		h2 { font-size: 20px; margin-top: 40px; }

		.card {
			background: white;
			border-radius: 12px;
			box-shadow: 0 4px 24px rgba(0,0,0,0.08);
			padding: 30px;
			margin-bottom: 30px;
		}

		/* --- Таблица анкет --- */
		table {
			width: 100%;
			border-collapse: collapse;
			font-size: 14px;
		}

		th {
			background: #1a1a2e;
			color: white;
			padding: 10px 12px;
			text-align: left;
			font-weight: 600;
		}

		td {
			padding: 10px 12px;
			border-bottom: 1px solid #e5e7eb;
			vertical-align: top;
		}

		tr:hover td { background: #f9fafb; }

		.lang-badge {
			display: inline-block;
			background: #e0e7ff;
			color: #3730a3;
			border-radius: 4px;
			padding: 2px 7px;
			font-size: 12px;
			margin: 2px;
		}

		/* --- Кнопки действий --- */
		.btn {
			display: inline-block;
			padding: 6px 12px;
			border-radius: 6px;
			font-size: 13px;
			font-weight: 600;
			cursor: pointer;
			border: none;
			text-decoration: none;
			font-family: inherit;
			transition: background 0.2s;
		}

		.btn-edit {
			background: #dbeafe;
			color: #1d4ed8;
		}

		.btn-edit:hover { background: #bfdbfe; }

		.btn-delete {
			background: #fee2e2;
			color: #dc2626;
			margin-left: 6px;
		}

		.btn-delete:hover { background: #fecaca; }

		.btn-save {
			background: #2c7be5;
			color: white;
			padding: 10px 24px;
			font-size: 15px;
		}

		.btn-save:hover { background: #1a5fc1; }

		.btn-cancel {
			background: #f3f4f6;
			color: #374151;
			padding: 10px 24px;
			font-size: 15px;
			margin-left: 10px;
		}

		/* --- Статистика --- */
		.stat-row {
			display: flex;
			align-items: center;
			margin-bottom: 10px;
			gap: 12px;
		}

		.stat-name {
			width: 120px;
			font-size: 14px;
			color: #374151;
			font-weight: 500;
		}

		.stat-bar-wrap {
			flex: 1;
			background: #e5e7eb;
			border-radius: 6px;
			height: 20px;
			overflow: hidden;
		}

		.stat-bar {
			height: 100%;
			background: linear-gradient(90deg, #2c7be5, #60a5fa);
			border-radius: 6px;
			transition: width 0.3s;
		}

		.stat-count {
			width: 30px;
			font-size: 14px;
			font-weight: 600;
			color: #1a1a2e;
			text-align: right;
		}

		/* --- Форма редактирования --- */
		.edit-form .field { margin-bottom: 16px; }

		.edit-form label {
			display: block;
			font-size: 13px;
			font-weight: 600;
			color: #555;
			margin-bottom: 5px;
			text-transform: uppercase;
			letter-spacing: 0.4px;
		}

		.edit-form input[type="text"],
		.edit-form input[type="tel"],
		.edit-form input[type="email"],
		.edit-form input[type="date"],
		.edit-form select,
		.edit-form textarea {
			width: 100%;
			padding: 9px 13px;
			border: 1.5px solid #d1d5db;
			border-radius: 8px;
			font-size: 14px;
			font-family: inherit;
			outline: none;
			background: #fafafa;
		}

		.edit-form select[multiple] { height: 140px; }
		.edit-form textarea { height: 90px; resize: vertical; }

		.edit-form .field-error input,
		.edit-form .field-error select,
		.edit-form .field-error textarea {
			border-color: #e53e3e;
			background: #fff5f5;
		}

		.edit-form .error-msg {
			font-size: 12px;
			color: #e53e3e;
			margin-top: 4px;
		}

		.edit-form .radio-group {
			display: flex;
			gap: 16px;
			margin-top: 4px;
		}

		.edit-form .radio-group label {
			display: flex;
			align-items: center;
			gap: 6px;
			font-weight: 400;
			text-transform: none;
			letter-spacing: 0;
			font-size: 14px;
		}

		.edit-form input[type="radio"],
		.edit-form input[type="checkbox"] {
			width: auto;
			accent-color: #2c7be5;
		}

		.topbar {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 24px;
		}

		.topbar a {
			font-size: 14px;
			color: #2c7be5;
			text-decoration: none;
		}
	</style>
</head>
<body>

<div class="topbar">
	<h1>🛠 Панель администратора</h1>
	<a href="../task6/index.cgi">← На главную</a>
</div>

{{if .EditID}}
	<!-- Режим редактирования -->
	<div class="card edit-form">
		<h2>✏️ Редактирование анкеты #{{.EditID}}</h2>
		<form action="admin.cgi?action=edit&id={{.EditID}}" method="POST">

			<input type="hidden" name="_csrf" value="{{$.CSRFToken}}">

			<div class="field {{if index .EditErrors "name"}}field-error{{end}}">
				<label>ФИО</label>
				<input type="text" name="name" value="{{.EditData.Name}}">
				{{if index .EditErrors "name"}}
					<div class="error-msg">{{index .EditErrors "name"}}</div>
				{{end}}
			</div>

			<div class="field {{if index .EditErrors "phone"}}field-error{{end}}">
				<label>Телефон</label>
				<input type="tel" name="phone" value="{{.EditData.Phone}}">
				{{if index .EditErrors "phone"}}
					<div class="error-msg">{{index .EditErrors "phone"}}</div>
				{{end}}
			</div>

			<div class="field {{if index .EditErrors "email"}}field-error{{end}}">
				<label>Email</label>
				<input type="email" name="email" value="{{.EditData.Email}}">
				{{if index .EditErrors "email"}}
					<div class="error-msg">{{index .EditErrors "email"}}</div>
				{{end}}
			</div>

			<div class="field {{if index .EditErrors "birthdate"}}field-error{{end}}">
				<label>Дата рождения</label>
				<input type="date" name="birthdate" value="{{.EditData.Birthdate}}">
				{{if index .EditErrors "birthdate"}}
					<div class="error-msg">{{index .EditErrors "birthdate"}}</div>
				{{end}}
			</div>

			<div class="field {{if index .EditErrors "gender"}}field-error{{end}}">
				<label>Пол</label>
				<div class="radio-group">
					<label>
						<input type="radio" name="gender" value="male"
							{{if eq .EditData.Gender "male"}}checked{{end}}> Мужской
					</label>
					<label>
						<input type="radio" name="gender" value="female"
							{{if eq .EditData.Gender "female"}}checked{{end}}> Женский
					</label>
				</div>
				{{if index .EditErrors "gender"}}
					<div class="error-msg">{{index .EditErrors "gender"}}</div>
				{{end}}
			</div>

			<div class="field {{if index .EditErrors "languages"}}field-error{{end}}">
				<label>Языки программирования</label>
				<select name="languages[]" multiple>
					{{range .Languages}}
					<option value="{{.ID}}"
						{{if $.IsSelectedEditLang .ID}}selected{{end}}>{{.Name}}</option>
					{{end}}
				</select>
				{{if index .EditErrors "languages"}}
					<div class="error-msg">{{index .EditErrors "languages"}}</div>
				{{end}}
			</div>

			<div class="field {{if index .EditErrors "bio"}}field-error{{end}}">
				<label>Биография</label>
				<textarea name="bio">{{.EditData.Bio}}</textarea>
				{{if index .EditErrors "bio"}}
					<div class="error-msg">{{index .EditErrors "bio"}}</div>
				{{end}}
			</div>

			<div class="field {{if index .EditErrors "contract"}}field-error{{end}}">
				<label style="display:flex;align-items:center;gap:8px;text-transform:none;letter-spacing:0;font-weight:400;font-size:14px;">
					<input type="checkbox" name="contract"
						{{if .EditData.Contract}}checked{{end}}> С контрактом ознакомлен(а)
				</label>
				{{if index .EditErrors "contract"}}
					<div class="error-msg">{{index .EditErrors "contract"}}</div>
				{{end}}
			</div>

			<button type="submit" class="btn btn-save">Сохранить</button>
			<a href="admin.cgi" class="btn btn-cancel">Отмена</a>
		</form>
	</div>

{{else}}
	<!-- Список анкет -->
	<div class="card">
		<h2>📋 Все анкеты ({{len .Applications}})</h2>
		{{if .Applications}}
		<table>
			<tr>
				<th>ID</th>
				<th>ФИО</th>
				<th>Телефон</th>
				<th>Email</th>
				<th>Дата рождения</th>
				<th>Пол</th>
				<th>Языки</th>
				<th>Действия</th>
			</tr>
			{{range .Applications}}
			<tr>
				<td>{{.ID}}</td>
				<td>{{.Name}}</td>
				<td>{{.Phone}}</td>
				<td>{{.Email}}</td>
				<td>{{.Birthdate}}</td>
				<td>{{if eq .Gender "male"}}Мужской{{else}}Женский{{end}}</td>
				<td>
					{{range .Languages}}
					<span class="lang-badge">{{.}}</span>
					{{end}}
				</td>
				<td>
					<a href="admin.cgi?action=edit&id={{.ID}}" class="btn btn-edit">✏️ Изменить</a>
					<form style="display:inline" action="admin.cgi?action=delete&id={{.ID}}" method="POST"
						onsubmit="return confirm('Удалить анкету #{{.ID}}?')">
						<input type="hidden" name="_csrf" value="{{$.CSRFToken}}">
						<button type="submit" class="btn btn-delete">🗑 Удалить</button>
					</form>
				</td>
			</tr>
			{{end}}
		</table>
		{{else}}
			<p style="color:#888">Анкет пока нет</p>
		{{end}}
	</div>

	<!-- Статистика -->
	<div class="card">
		<h2>📊 Статистика по языкам</h2>
		{{$max := 1}}
		{{range .Stats}}{{if gt .Count $max}}{{$max = .Count}}{{end}}{{end}}
		{{range .Stats}}
		<div class="stat-row">
			<span class="stat-name">{{.Name}}</span>
			<div class="stat-bar-wrap">
				<div class="stat-bar" style="width: {{if $max}}{{mul .Count 100 | div $max}}%{{else}}0%{{end}}"></div>
			</div>
			<span class="stat-count">{{.Count}}</span>
		</div>
		{{end}}
	</div>
{{end}}

</body>
</html>`
