package common

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
)

func CheckKey(request *http.Request, key string) bool {
	requestKey := request.Header.Get("Key")
	return subtle.ConstantTimeCompare([]byte(key), []byte(requestKey)) == 1
}

func CheckToken(token string, requestToken string) bool {
	if token == "" || requestToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(requestToken)) == 1
}

func GenerateSessionToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("unable to generate session token: %w", err)
	}
	return hex.EncodeToString(raw), nil
}