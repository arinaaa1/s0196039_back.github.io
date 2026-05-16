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
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, Helvetica, sans-serif;
            background: linear-gradient(145deg, #ffe4ec 0%, #ffd6e2 100%);
            min-height: 100vh;
            padding: 40px 20px;
        }

        .topbar {
            max-width: 680px;
            margin: 0 auto 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: rgba(255, 255, 255, 0.7);
            backdrop-filter: blur(4px);
            padding: 0.8rem 1.5rem;
            border-radius: 2rem;
            border: 1px solid rgba(255, 245, 245, 0.8);
        }

        .topbar-user {
            font-size: 0.9rem;
            color: #b34e72;
        }

        .topbar-user strong {
            color: #c4456c;
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
            text-decoration: none;
        }

        .card {
            background: rgba(255, 255, 255, 0.88);
            backdrop-filter: blur(4px);
            border-radius: 2rem;
            box-shadow: 0 20px 35px -12px rgba(236, 72, 153, 0.12), 0 0 0 1px rgba(255, 245, 245, 0.7) inset;
            padding: 2.2rem 2rem 2.5rem;
            width: 100%;
            max-width: 680px;
            margin: 0 auto;
            transition: all 0.2s ease;
        }

        h1 {
            font-size: 2rem;
            font-weight: 500;
            letter-spacing: -0.01em;
            background: linear-gradient(135deg, #c4456c, #b8315a);
            background-clip: text;
            -webkit-background-clip: text;
            color: transparent;
            margin-bottom: 1.8rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid rgba(200, 70, 110, 0.2);
            display: inline-block;
        }

        .field {
            margin-bottom: 1.5rem;
        }

        .field > label {
            display: block;
            font-size: 0.9rem;
            font-weight: 500;
            color: #9e4466;
            margin-bottom: 0.4rem;
            letter-spacing: -0.2px;
        }

        input[type="text"],
        input[type="tel"],
        input[type="email"],
        input[type="date"],
        select,
        textarea {
            width: 100%;
            padding: 0.8rem 1rem;
            border: 1.5px solid #f3cdd8;
            border-radius: 1.2rem;
            font-size: 0.95rem;
            font-family: inherit;
            background: #ffffffdd;
            transition: all 0.2s ease;
            outline: none;
            color: #2d2a2b;
        }

        input:focus,
        select:focus,
        textarea:focus {
            border-color: #e07c9e;
            box-shadow: 0 0 0 3px rgba(224, 124, 158, 0.2);
            background: #fff;
        }

        .field-error input,
        .field-error select,
        .field-error textarea {
            border-color: #e86c8c;
            background: #fff5f7;
        }

        .field-error input:focus,
        .field-error select:focus,
        .field-error textarea:focus {
            border-color: #d9537a;
        }

        .error-msg {
            font-size: 0.75rem;
            color: #d94a73;
            margin-top: 0.3rem;
            margin-left: 0.5rem;
            display: flex;
            align-items: center;
            gap: 4px;
        }

        .error-msg::before {
            content: "✦";
            display: inline-flex;
            align-items: center;
            justify-content: center;
            font-size: 0.7rem;
            color: #e85d8b;
            flex-shrink: 0;
        }

        textarea {
            height: 110px;
            resize: vertical;
        }

        select[multiple] {
            height: 160px;
            padding: 0.6rem;
            border-radius: 1rem;
            background: #ffffffdd;
        }

        select[multiple] option {
            padding: 0.4rem 0.6rem;
            border-radius: 0.8rem;
            margin: 2px 0;
        }

        select[multiple] option:checked {
            background: #fbc1d2 linear-gradient(0deg, #f7a9c0 0%, #f7a9c0 100%);
            color: #4a1e2f;
        }

        .radio-group {
            display: flex;
            gap: 1.5rem;
            flex-wrap: wrap;
            align-items: center;
            margin-top: 0.3rem;
        }

        .radio-group label,
        .checkbox-label {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.95rem;
            font-weight: 400;
            color: #a65472;
            cursor: pointer;
        }

        input[type="radio"],
        input[type="checkbox"] {
            accent-color: #e36c92;
            width: 1rem;
            height: 1rem;
            margin: 0;
            cursor: pointer;
        }

        .btn {
            width: 100%;
            padding: 0.9rem;
            background: linear-gradient(95deg, #e47297, #d95580);
            color: white;
            border: none;
            border-radius: 2rem;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.25s ease;
            margin-top: 0.6rem;
            box-shadow: 0 4px 10px rgba(217, 85, 128, 0.2);
            letter-spacing: 0.3px;
        }

        .btn:hover {
            background: linear-gradient(95deg, #dc5f88, #c9456f);
            transform: scale(0.98);
            box-shadow: 0 6px 14px rgba(217, 85, 128, 0.25);
        }

        .success-banner {
            background: #fff0f3;
            border: 1.5px solid #e38aa8;
            border-radius: 1.2rem;
            padding: 0.9rem 1.2rem;
            color: #b6436a;
            font-size: 0.9rem;
            margin-bottom: 1.8rem;
            text-align: center;
            font-weight: 500;
        }

        @media (max-width: 600px) {
            body {
                padding: 1.2rem;
            }
            .card, .topbar {
                padding: 1.6rem;
            }
            h1 {
                font-size: 1.8rem;
            }
            .radio-group {
                flex-direction: column;
                align-items: flex-start;
                gap: 0.6rem;
            }
        }

        .card {
            animation: fadeSlideUp 0.45s ease-out;
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
