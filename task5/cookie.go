package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

func setSessionCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
	})
}

func setPersistentCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
	})
}

func deleteCookies(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
}

func getCookieValue(r *http.Request, name string) (string, bool) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

func encodeToCookie(v any) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return url.QueryEscape(string(jsonBytes)), nil
}

func decodeFromCookie(s string, v any) error {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(decoded), v)
}

func saveErrorsToCookie(w http.ResponseWriter, data FormData, errors FormErrors) {
	if encoded, err := encodeToCookie(errors); err == nil {
		setSessionCookie(w, "form_errors", encoded)
	}
	if encoded, err := encodeToCookie(data); err == nil {
		setSessionCookie(w, "form_values", encoded)
	}
}

func saveSuccessToCookie(w http.ResponseWriter, data FormData) {
	if encoded, err := encodeToCookie(data); err == nil {
		setPersistentCookie(w, "form_values", encoded)
	}
	setSessionCookie(w, "form_success", "1")
}

func loadFromCookies(w http.ResponseWriter, r *http.Request) PageData {
	var data PageData
	if raw, ok := getCookieValue(r, "form_values"); ok {
		decodeFromCookie(raw, &data.Values)
	}

	if raw, ok := getCookieValue(r, "form_errors"); ok {
		data.Errors = make(FormErrors)
		decodeFromCookie(raw, &data.Errors)
		deleteCookies(w, "form_errors")
	}
	if _, ok := getCookieValue(r, "form_success"); ok {
		data.Success = true
		deleteCookies(w, "form_success")
	}
	return data
}
