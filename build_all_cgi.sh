#!/bin/bash
set -e

# Настройки кросс-компиляции (под твой сервер)
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
LDFLAGS="-s -w"

echo "=============================="
echo "=== task3: form.cgi       ==="
echo "=============================="
cd task3
go build -ldflags="$LDFLAGS" -o form.cgi form.go
cd ..

echo "=============================="
echo "=== task4: form.cgi       ==="
echo "=============================="
cd task4
go build -ldflags="$LDFLAGS" -o form.cgi *.go
cd ..

echo "=============================="
echo "=== task5: мультиплексный ==="
echo "=============================="
cd task5
# Собираем ОДИН бинарник
go build -ldflags="$LDFLAGS" -o app.cgi *.go
# Создаём (или обновляем) симлинки с нужными именами
for name in form login edit logout index; do
    ln -sf app.cgi ${name}.cgi
done
cd ..

echo "=============================="
echo "=== task6: мультиплексный ==="
echo "=============================="
cd task6
go build -ldflags="$LDFLAGS" -o app.cgi *.go
for name in form login edit logout index admin; do
    ln -sf app.cgi ${name}.cgi
done
cd ..

echo "=============================="
echo "=== task8: api.cgi        ==="
echo "=============================="
cd task8
go build -ldflags="$LDFLAGS" -o api.cgi *.go
cd ..

echo ""
echo "✅ Все CGI успешно собраны!"