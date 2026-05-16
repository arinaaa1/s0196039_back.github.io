#!/bin/bash
export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
LDFLAGS="-s -w"
set -e

echo "=== task3 ==="
cd task3 && go build -ldflags="$LDFLAGS" -o form.cgi form.go && cd ..

echo "=== task4 ==="
cd task4 && go build -ldflags="$LDFLAGS" -o form.cgi *.go && cd ..

echo "=== task5 ==="
cd task5 && go build -ldflags="$LDFLAGS" -o app.cgi *.go && for n in form login edit logout index; do ln -sf app.cgi $n.cgi; done && cd ..

echo "=== task6 ==="
cd task6 && go build -ldflags="$LDFLAGS" -o app.cgi *.go && for n in form login edit logout index admin; do ln -sf app.cgi $n.cgi; done && cd ..

echo "=== task8 ==="
cd task8 && go build -ldflags="$LDFLAGS" -o api.cgi *.go && cd ..

echo "=== Done! ==="
