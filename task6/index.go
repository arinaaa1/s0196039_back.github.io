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
		* { box-sizing: border-box; margin: 0; padding: 0; }

		body {
			font-family: Arial, sans-serif;
			background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
		}

		.card {
			background: rgba(255,255,255,0.05);
			backdrop-filter: blur(10px);
			border: 1px solid rgba(255,255,255,0.1);
			border-radius: 20px;
			padding: 60px 50px;
			width: 100%;
			max-width: 480px;
			text-align: center;
		}

		.logo {
			font-size: 72px;
			margin-bottom: 16px;
			display: block;
		}

		h1 {
			font-size: 28px;
			font-weight: 700;
			color: #ffffff;
			margin-bottom: 8px;
		}

		.subtitle {
			font-size: 15px;
			color: rgba(255,255,255,0.5);
			margin-bottom: 40px;
		}

		.user-greeting {
			background: rgba(44,123,229,0.2);
			border: 1px solid rgba(44,123,229,0.4);
			border-radius: 10px;
			padding: 12px 20px;
			color: #7eb8f7;
			font-size: 14px;
			margin-bottom: 24px;
		}

		.user-greeting strong {
			color: #ffffff;
		}

		.btn {
			display: block;
			width: 100%;
			padding: 14px;
			border-radius: 10px;
			font-size: 16px;
			font-weight: 600;
			cursor: pointer;
			text-decoration: none;
			transition: all 0.2s;
			border: none;
			margin-bottom: 12px;
			font-family: inherit;
		}

		.btn-primary {
			background: #2c7be5;
			color: white;
		}

		.btn-primary:hover { background: #1a5fc1; }

		.btn-secondary {
			background: rgba(255,255,255,0.08);
			color: rgba(255,255,255,0.8);
			border: 1px solid rgba(255,255,255,0.15);
		}

		.btn-secondary:hover {
			background: rgba(255,255,255,0.15);
			color: white;
		}

		.btn-danger {
			background: rgba(229,62,62,0.15);
			color: #fc8181;
			border: 1px solid rgba(229,62,62,0.3);
		}

		.btn-danger:hover {
			background: rgba(229,62,62,0.25);
		}

		.divider {
			height: 1px;
			background: rgba(255,255,255,0.1);
			margin: 8px 0 20px;
		}

		.btn-admin {
   			background: rgba(255,255,255,0.05);
    		color: rgba(255,255,255,0.4);
    		border: 1px solid rgba(255,255,255,0.08);
    		font-size: 13px;
    		padding: 10px;
		}
		.btn-admin:hover {
			background: rgba(255,255,255,0.1);
			color: rgba(255,255,255,0.7);
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
