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
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, Helvetica, sans-serif;
            background: linear-gradient(145deg, #ffe4ec 0%, #ffd6e2 100%);
            padding: 30px 20px;
        }

        h1, h2 {
            margin-bottom: 20px;
        }

        h1 {
            font-size: 1.8rem;
            font-weight: 500;
            letter-spacing: -0.01em;
            background: linear-gradient(135deg, #c4456c, #b8315a);
            background-clip: text;
            -webkit-background-clip: text;
            color: transparent;
        }

        h2 {
            font-size: 1.4rem;
            color: #b34e72;
            font-weight: 500;
        }

        .card {
            background: rgba(255, 255, 255, 0.88);
            backdrop-filter: blur(4px);
            border-radius: 2rem;
            box-shadow: 0 20px 35px -12px rgba(236, 72, 153, 0.12), 0 0 0 1px rgba(255, 245, 245, 0.7) inset;
            padding: 1.8rem;
            margin-bottom: 2rem;
            transition: all 0.2s ease;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.85rem;
        }

        th {
            background: #e8a7bb;
            color: #4a1e2f;
            padding: 10px 12px;
            text-align: left;
            font-weight: 600;
            border-radius: 1rem 1rem 0 0;
        }

        td {
            padding: 10px 12px;
            border-bottom: 1px solid rgba(217, 85, 128, 0.2);
            vertical-align: top;
            color: #2d2a2b;
        }

        tr:hover td {
            background: rgba(255, 235, 240, 0.6);
        }

        .lang-badge {
            display: inline-block;
            background: #fce4ec;
            color: #c4456c;
            border-radius: 2rem;
            padding: 2px 10px;
            font-size: 0.7rem;
            margin: 2px;
            font-weight: 500;
        }

        .btn {
            display: inline-block;
            padding: 6px 14px;
            border-radius: 2rem;
            font-size: 0.75rem;
            font-weight: 500;
            cursor: pointer;
            border: none;
            text-decoration: none;
            font-family: inherit;
            transition: all 0.2s;
        }

        .btn-edit {
            background: #fce4ec;
            color: #b8315a;
        }

        .btn-edit:hover {
            background: #f8cddb;
            transform: scale(0.96);
        }

        .btn-delete {
            background: #ffe4e4;
            color: #d94a73;
            margin-left: 6px;
        }

        .btn-delete:hover {
            background: #ffd0d0;
            transform: scale(0.96);
        }

        .btn-save {
            background: linear-gradient(95deg, #e47297, #d95580);
            color: white;
            padding: 10px 24px;
            font-size: 0.9rem;
            box-shadow: 0 2px 8px rgba(217, 85, 128, 0.2);
        }

        .btn-save:hover {
            background: linear-gradient(95deg, #dc5f88, #c9456f);
            transform: scale(0.98);
        }

        .btn-cancel {
            background: rgba(217, 85, 128, 0.1);
            color: #b34e72;
            padding: 10px 24px;
            font-size: 0.9rem;
            margin-left: 10px;
            border: 1px solid rgba(217, 85, 128, 0.3);
        }

        .btn-cancel:hover {
            background: rgba(217, 85, 128, 0.2);
        }

        .stat-row {
            display: flex;
            align-items: center;
            margin-bottom: 12px;
            gap: 12px;
        }

        .stat-name {
            width: 120px;
            font-size: 0.85rem;
            color: #b34e72;
            font-weight: 500;
        }

        .stat-bar-wrap {
            flex: 1;
            background: #f3cdd8;
            border-radius: 1rem;
            height: 20px;
            overflow: hidden;
        }

        .stat-bar {
            height: 100%;
            background: linear-gradient(90deg, #e47297, #d95580);
            border-radius: 1rem;
            transition: width 0.3s;
        }

        .stat-count {
            width: 35px;
            font-size: 0.85rem;
            font-weight: 600;
            color: #b8315a;
            text-align: right;
        }

        .edit-form .field {
            margin-bottom: 1.2rem;
        }

        .edit-form label {
            display: block;
            font-size: 0.8rem;
            font-weight: 500;
            color: #9e4466;
            margin-bottom: 0.3rem;
            letter-spacing: -0.2px;
        }

        .edit-form input[type="text"],
        .edit-form input[type="tel"],
        .edit-form input[type="email"],
        .edit-form input[type="date"],
        .edit-form select,
        .edit-form textarea {
            width: 100%;
            padding: 0.7rem 1rem;
            border: 1.5px solid #f3cdd8;
            border-radius: 1.2rem;
            font-size: 0.9rem;
            font-family: inherit;
            background: #ffffffdd;
            transition: all 0.2s;
            outline: none;
        }

        .edit-form input:focus,
        .edit-form select:focus,
        .edit-form textarea:focus {
            border-color: #e07c9e;
            box-shadow: 0 0 0 3px rgba(224, 124, 158, 0.2);
            background: #fff;
        }

        .edit-form select[multiple] {
            height: 140px;
            padding: 0.5rem;
        }

        .edit-form textarea {
            height: 90px;
            resize: vertical;
        }

        .edit-form .field-error input,
        .edit-form .field-error select,
        .edit-form .field-error textarea {
            border-color: #e86c8c;
            background: #fff5f7;
        }

        .edit-form .error-msg {
            font-size: 0.7rem;
            color: #d94a73;
            margin-top: 0.3rem;
            margin-left: 0.5rem;
        }

        .edit-form .radio-group {
            display: flex;
            gap: 1.5rem;
            margin-top: 0.3rem;
        }

        .edit-form .radio-group label {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-weight: 400;
            font-size: 0.9rem;
            color: #a65472;
            text-transform: none;
            letter-spacing: 0;
        }

        .edit-form input[type="radio"],
        .edit-form input[type="checkbox"] {
            accent-color: #e36c92;
            width: 1rem;
            height: 1rem;
            margin: 0;
        }

        .topbar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
            background: rgba(255, 255, 255, 0.7);
            backdrop-filter: blur(4px);
            padding: 0.8rem 1.5rem;
            border-radius: 2rem;
            border: 1px solid rgba(255, 245, 245, 0.8);
        }

        .topbar a {
            font-size: 0.85rem;
            color: #d95580;
            text-decoration: none;
            font-weight: 500;
            padding: 0.3rem 0.8rem;
            border-radius: 2rem;
            transition: background 0.2s;
        }

        .topbar a:hover {
            background: rgba(217, 85, 128, 0.1);
        }

        @keyframes fadeSlideUp {
            from {
                opacity: 0;
                transform: translateY(18px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        .card {
            animation: fadeSlideUp 0.45s ease-out;
        }
    </style>
</head>
<body>

<div class="topbar">
    <h1>🛠 Панель администратора</h1>
    <a href="../task6/index.cgi">← На главную</a>
</div>

{{if .EditID}}
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
                <label style="display:flex;align-items:center;gap:8px;font-weight:400;font-size:0.9rem;">
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
    <div class="card">
        <h2>📋 Все анкеты ({{len .Applications}})</h2>
        {{if .Applications}}
        <table>
            <thead>
                <tr><th>ID</th><th>ФИО</th><th>Телефон</th><th>Email</th><th>Дата рождения</th><th>Пол</th><th>Языки</th><th>Действия</th></tr>
            </thead>
            <tbody>
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
            </tbody>
        </table>
        {{else}}
            <p style="color:#b35f7c">Анкет пока нет</p>
        {{end}}
    </div>

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
