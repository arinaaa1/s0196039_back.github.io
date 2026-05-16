package main

import (
	"net/http"
	"net/http/cgi"
)

func runLogout() {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deleteJWTCookie(w)
		http.Redirect(w, r, "index.cgi", http.StatusFound)
	}))
}
