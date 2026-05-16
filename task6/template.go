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
            font-family: Arial, sans-serif;
            background: #f0f2f5;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 40px 20px;
        }

        .card {
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 24px rgba(0,0,0,0.10);
            padding: 40px;
            width: 100%;
            max-width: 560px;
        }

        h1 {
            font-size: 24px;
            font-weight: 700;
            color: #1a1a2e;
            margin-bottom: 28px;
            text-align: center;
            letter-spacing: -0.5px;
        }

        .field {
            margin-bottom: 20px;
        }

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

        input:focus,
        select:focus,
        textarea:focus {
            border-color: #2c7be5;
            background: #fff;
        }

        .field-error input,
        .field-error select,
        .field-error textarea {
            border-color: #e53e3e;
            background: #fff5f5;
        }

        .field-error input:focus,
        .field-error select:focus,
        .field-error textarea:focus {
            border-color: #c53030;
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

        textarea {
            height: 110px;
            resize: vertical;
        }

        select[multiple] {
            height: 160px;
            padding: 6px;
        }

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
            transition: background 0.2s, transform 0.1s;
            letter-spacing: 0.2px;
        }

        .btn:hover {
            background: #1a5fc1;
        }

        .btn:active {
            transform: scale(0.99);
        }

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

        .credentials-banner {
            background: #fffbeb;
            border: 1.5px solid #f6ad55;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 24px;
        }

        .credentials-banner h3 {
            font-size: 16px;
            color: #744210;
            margin-bottom: 8px;
        }

        .credentials-banner p {
            font-size: 13px;
            color: #975a16;
            margin-bottom: 12px;
        }

        .cred-row {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 6px;
            font-size: 15px;
        }

        .cred-label {
            color: #744210;
            font-size: 13px;
            width: 60px;
        }

        .cred-row strong {
            font-family: monospace;
            font-size: 16px;
            background: #fef3c7;
            padding: 2px 8px;
            border-radius: 4px;
            letter-spacing: 0.5px;
        }

        .btn-login {
            display: inline-block;
            margin-top: 12px;
            padding: 8px 16px;
            background: #2c7be5;
            color: white;
            border-radius: 6px;
            text-decoration: none;
            font-size: 14px;
            font-weight: 600;
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
