package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)


func HashPassword(password string) (string, error){
	// Use argon2id to hash the password
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password string, hash string) (bool, error){
	// Use argon2id to compare the password with the hash
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error){
	var claims jwt.RegisteredClaims = jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		Subject: userID.String(),
	}
	var token *jwt.Token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error){
	var claims *jwt.RegisteredClaims = &jwt.RegisteredClaims{}
	var keyFunc jwt.Keyfunc = func(token *jwt.Token) (interface{}, error){
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte(tokenSecret), nil
	}
	
	_, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Error validating JWT: %w", err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Invalid user ID in token claims: %w", err)
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error){
	var authHeader string = headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is missing")
	}
	
	i, err := fmt.Sscanf(authHeader, "Bearer %s", &authHeader)
	if err != nil || i != 1 {
		return "", fmt.Errorf("Invalid Authorization header format")
	}

	return authHeader, nil
}

func MakeRefreshToken() string{
	var key []byte = make([]byte, 32)
	rand.Read(key)

	return hex.EncodeToString(key)
}

