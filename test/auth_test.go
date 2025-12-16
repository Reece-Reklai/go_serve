package test

import (
	"strings"
	"testing"
	"time"

	"github.com/Reece-Reklai/go_serve/internal/auth"
	"github.com/google/uuid"
)

// Test Conditions (JWT):
// 1) invalid signed string
// 2) invalid token secret
// 3) token expires
func TestAuthJWT(t *testing.T) {
	tokenSecret := "hello"
	timeExpiredIn := time.Duration(1) * time.Second
	signedString, err := auth.MakeJWT(uuid.New(), tokenSecret, timeExpiredIn)
	if err != nil {
		t.Errorf(`Json Web token was not created: %v\n`, err)
	}
	// Purposely Fail Test Conditions Below:
	// invalidSignedString := "random"
	// invalidTokenSecret := "random"
	// time.Sleep(1 * time.Second)
	_, err = auth.ValidateJWT(signedString, tokenSecret)

	if err != nil {
		t.Errorf(`Validation Failed: %v\n`, err)
	}
}

func TestStripStringOnHeaderAuth(t *testing.T) {
	bearerToken := "Bearer TOKEN_STRING"
	strip := strings.Split(bearerToken, " ")
	if strip[1] != "TOKEN_STRING" {
		t.Error("Word was not found")
	}
}
