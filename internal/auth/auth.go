package auth

import (
	"github.com/alexedwards/argon2id"
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

