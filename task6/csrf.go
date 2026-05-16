package main

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

const csrfCookieName = "csrf_token"
const csrfFieldName = "_csrf"

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func getOrCreateCSRFToken(w http.ResponseWriter, r *http.Request) string {
	if token, ok := getCookieValue(r, csrfCookieName); ok {
		return token
	}
	token, err := generateCSRFToken()
	if err != nil {
		return ""
	}
	setSessionCookie(w, csrfCookieName, token)
	return token
}

func validateCSRFToken(r *http.Request) bool {
	cookieToken, ok := getCookieValue(r, csrfCookieName)
	if !ok || cookieToken == "" {
		return false
	}
	formToken := r.FormValue(csrfFieldName)
	return hmac.Equal([]byte(cookieToken), []byte(formToken))
}
