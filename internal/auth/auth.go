package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "failed to hash", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return match, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	mySigningKey := []byte(tokenSecret)

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy",
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		return "something went wrong", err
	}
	return ss, nil
}
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	customClaims := struct {
		jwt.RegisteredClaims
	}{}
	token, err := jwt.ParseWithClaims(tokenString, &customClaims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	stringToUUID, err := uuid.ParseBytes([]byte(id))
	if err != nil {
		return uuid.Nil, err
	}
	return stringToUUID, err
}

func GetBearerToken(headers http.Header) (string, error) {
	headerBearer := headers.Get("Authorization")
	if headerBearer == "" {
		return "", errors.New("empty header authorization")
	}
	strip := strings.Split(headerBearer, " ")
	return strip[1], nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)
	hex_encoded := hex.EncodeToString(key)
	return hex_encoded, nil
}
