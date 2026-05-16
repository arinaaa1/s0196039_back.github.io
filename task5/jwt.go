package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const jwtSecret = "it-is-the-most-top-secret-key-from-Epstein's-files"

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

type JWTPayload struct {
	ApplicationID int64  `json:"application_id"`
	Login         string `json:"login"`
	Exp           int64  `json:"exp"`
}

func generateJWT(applicationID int64, login string) (string, error) {
	payload := JWTPayload{
		ApplicationID: applicationID,
		Login:         login,
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("generateJWT marshal: %w", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := jwtHeader + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(jwtSecret))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + signature, nil
}

func validateJWT(token string) (JWTPayload, error) {
	var payload JWTPayload

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return payload, errors.New("Invalid token construction")
	}
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(jwtSecret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum((nil)))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return payload, errors.New("Invalid token signature")
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return payload, fmt.Errorf("decode payload: %w", err)
	}

	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return payload, fmt.Errorf("unmarshal payload: %w", err)
	}

	if time.Now().Unix() > payload.Exp {
		return payload, errors.New("Token has expired")
	}

	return payload, nil
}

func getJWTFromCookie(r *http.Request) (JWTPayload, bool) {
	token, ok := getCookieValue(r, "jwt_token")
	if !ok {
		return JWTPayload{}, false
	}
	payload, err := validateJWT(token)
	if err != nil {
		return JWTPayload{}, false
	}
	return payload, true
}

func setJWTCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func deleteJWTCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
}
