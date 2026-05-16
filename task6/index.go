package main

import (
	"html/template"
	"net/http"
	"net/http/cgi"
)

var indexTemplate = template.Must(template.New("index").Parse(indexHTML))

type IndexPageData struct {
	IsLoggedIn bool
	Login      string
}

func runIndex() {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := IndexPageData{}
		if payload, ok := getJWTFromCookie(r); ok {
			data.IsLoggedIn = true
			data.Login = payload.Login
		}

		indexTemplate.Execute(w, data)

	}))
}

const indexHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Анкета — Главная</title>
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
            padding: 2rem;
        }

        .card {
            background: rgba(255, 255, 255, 0.88);
            backdrop-filter: blur(4px);
            border-radius: 2rem;
            box-shadow: 0 20px 35px -12px rgba(236, 72, 153, 0.12), 0 0 0 1px rgba(255, 245, 245, 0.7) inset;
            padding: 2.5rem 2rem;
            width: 100%;
            max-width: 480px;
            text-align: center;
            transition: all 0.2s ease;
            animation: fadeSlideUp 0.45s ease-out;
        }

        .logo {
            font-size: 4rem;
            margin-bottom: 1rem;
            display: block;
        }

        h1 {
            font-size: 2rem;
            font-weight: 500;
            letter-spacing: -0.01em;
            background: linear-gradient(135deg, #c4456c, #b8315a);
            background-clip: text;
            -webkit-background-clip: text;
            color: transparent;
            margin-bottom: 0.5rem;
        }

        .subtitle {
            font-size: 0.9rem;
            color: #b35f7c;
            margin-bottom: 2rem;
        }

        .user-greeting {
            background: rgba(224, 124, 158, 0.1);
            border: 1px solid rgba(224, 124, 158, 0.3);
            border-radius: 1.5rem;
            padding: 0.7rem 1rem;
            color: #b34e72;
            font-size: 0.85rem;
            margin-bottom: 1.5rem;
        }

        .user-greeting strong {
            color: #c4456c;
        }

        .btn {
            display: block;
            width: 100%;
            padding: 0.8rem;
            border-radius: 2rem;
            font-size: 0.95rem;
            font-weight: 500;
            cursor: pointer;
            text-decoration: none;
            transition: all 0.2s;
            border: none;
            margin-bottom: 0.8rem;
            font-family: inherit;
        }

        .btn-primary {
            background: linear-gradient(95deg, #e47297, #d95580);
            color: white;
            box-shadow: 0 4px 10px rgba(217, 85, 128, 0.2);
        }

        .btn-primary:hover {
            background: linear-gradient(95deg, #dc5f88, #c9456f);
            transform: scale(0.98);
        }

        .btn-secondary {
            background: rgba(217, 85, 128, 0.1);
            color: #c4456c;
            border: 1px solid rgba(217, 85, 128, 0.3);
        }

        .btn-secondary:hover {
            background: rgba(217, 85, 128, 0.2);
        }

        .btn-danger {
            background: rgba(229, 62, 62, 0.1);
            color: #d94a73;
            border: 1px solid rgba(217, 85, 128, 0.3);
        }

        .btn-danger:hover {
            background: rgba(229, 62, 62, 0.2);
        }

        .btn-admin {
            background: rgba(217, 85, 128, 0.05);
            color: #b34e72;
            border: 1px solid rgba(217, 85, 128, 0.2);
            font-size: 0.85rem;
            padding: 0.7rem;
        }

        .btn-admin:hover {
            background: rgba(217, 85, 128, 0.15);
        }

        .divider {
            height: 1px;
            background: rgba(217, 85, 128, 0.2);
            margin: 0.5rem 0 1.2rem;
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
<div class="card">
    <span class="logo">📋</span>
    <h1>Система анкетирования</h1>
    <p class="subtitle">Заполните анкету или войдите чтобы изменить данные</p>

    {{if .IsLoggedIn}}
        <div class="user-greeting">
            Вы вошли как <strong>{{.Login}}</strong>
        </div>
        <a href="edit.cgi" class="btn btn-primary">✏️ Редактировать анкету</a>
        <div class="divider"></div>
        <a href="logout.cgi" class="btn btn-danger">Выйти</a>
    {{else}}
        <a href="form.cgi" class="btn btn-primary">📝 Заполнить анкету</a>
        <a href="login.cgi" class="btn btn-secondary">🔐 Войти</a>
        <div class="divider"></div>
        <a href="admin.cgi" class="btn btn-admin">🛠 Войти как администратор</a>
    {{end}}
</div>
</body>
</html>`
