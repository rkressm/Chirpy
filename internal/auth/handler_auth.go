package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := strings.Replace(headers.Get("Authorization"), "ApiKey ", "", 1)
	if apiKey == "" {
		return "", fmt.Errorf("apikey missing")
	}
	return apiKey, nil
}

func MakeRefresherToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	token := hex.EncodeToString(key)
	return token
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := strings.Replace(headers.Get("Authorization"), "Bearer ", "", 1)
	if tokenString == "" {
		return "", fmt.Errorf("Token missing")
	}
	return tokenString, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	jwtToken, err := newToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return jwtToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsingwithclaims: %w", err)
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("error getting subject: %w", err)
	}
	value, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error uuid parsing: %w", err)
	}
	return value, nil
}

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("error comparing hashes: %w", err)
	}
	return match, nil
}
