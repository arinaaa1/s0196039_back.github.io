package main

import (
	"html/template"
	"net/http"
)

type Language struct {
	ID   string
	Name string
}

var allLanguages = []Language{
	{"1", "Pascal"}, {"2", "C"}, {"3", "C++"},
	{"4", "JavaScript"}, {"5", "PHP"}, {"6", "Python"},
	{"7", "Java"}, {"8", "Haskell"}, {"9", "Clojure"},
	{"10", "Prolog"}, {"11", "Scala"}, {"12", "Go"},
}

type FormData struct {
	Name      string
	Phone     string
	Email     string
	Birthdate string
	Gender    string
	Bio       string
	Languages []string
	Contract  bool
}

type FormErrors map[string]string

type PageData struct {
	Values    FormData
	Errors    FormErrors
	Languages []Language
	Success   bool
	CSRFToken string
}

func (p PageData) IsSelectedLang(id string) bool {
	for _, selected := range p.Values.Languages {
		if selected == id {
			return true
		}
	}
	return false
}

var tmpl = template.Must(template.New("form").Parse(formHTML))

func renderForm(w http.ResponseWriter, data PageData, creds map[string]string) {
	data.Languages = allLanguages
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, struct {
		PageData
		NewCreds map[string]string
	}{data, creds})
}

const formHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Анкета</title>
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
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 40px 20px;
        }

        .card {
            background: rgba(255, 255, 255, 0.88);
            backdrop-filter: blur(4px);
            border-radius: 2rem;
            box-shadow: 0 20px 35px -12px rgba(236, 72, 153, 0.12), 0 0 0 1px rgba(255, 245, 245, 0.7) inset;
            padding: 2.2rem 2rem 2.5rem;
            width: 100%;
            max-width: 680px;
            transition: all 0.2s ease;
            animation: fadeSlideUp 0.45s ease-out;
        }

        h1 {
            font-size: 2.1rem;
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

        .credentials-banner {
            background: #fff0f3;
            border: 1.5px solid #e38aa8;
            border-radius: 1.2rem;
            padding: 1.2rem;
            margin-top: 1.5rem;
        }

        .credentials-banner h3 {
            font-size: 1rem;
            color: #b6436a;
            margin-bottom: 0.5rem;
        }

        .credentials-banner p {
            font-size: 0.8rem;
            color: #b35f7c;
            margin-bottom: 0.8rem;
        }

        .cred-row {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            margin-bottom: 0.5rem;
            font-size: 0.9rem;
        }

        .cred-label {
            color: #9e4466;
            font-size: 0.8rem;
            width: 60px;
        }

        .cred-row strong {
            font-family: monospace;
            font-size: 0.9rem;
            background: #ffe4ec;
            padding: 0.2rem 0.6rem;
            border-radius: 0.8rem;
            letter-spacing: 0.5px;
            color: #c4456c;
        }

        .btn-login {
            display: inline-block;
            margin-top: 0.8rem;
            padding: 0.4rem 1.2rem;
            background: linear-gradient(95deg, #e47297, #d95580);
            color: white;
            border-radius: 2rem;
            text-decoration: none;
            font-size: 0.8rem;
            font-weight: 500;
            transition: all 0.2s;
        }

        .btn-login:hover {
            background: linear-gradient(95deg, #dc5f88, #c9456f);
            transform: scale(0.96);
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

        @media (max-width: 600px) {
            body {
                padding: 1.2rem;
            }
            .card {
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
    </style>
</head>
<body>
<div class="card">
    <h1>Анкета</h1>

    {{if .Success}}
    <div class="success-banner">✅ Анкета успешно сохранена!</div>
    {{end}}

    <form action="form.cgi" method="POST">
        <input type="hidden" name="_csrf" value="{{.CSRFToken}}">

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

        <button type="submit" class="btn">Сохранить</button>

        {{if .NewCreds}}
        <div class="credentials-banner">
            <h3>🎉 Анкета отправлена!</h3>
            <p>Сохраните данные для входа — они показываются только один раз:</p>
            <div class="cred-row">
                <span class="cred-label">Логин:</span>
                <strong>{{index .NewCreds "login"}}</strong>
            </div>
            <div class="cred-row">
                <span class="cred-label">Пароль:</span>
                <strong>{{index .NewCreds "password"}}</strong>
            </div>
            <a href="login.cgi" class="btn-login">Войти →</a>
        </div>
        {{end}}

    </form>
</div>
</body>
</html>`
