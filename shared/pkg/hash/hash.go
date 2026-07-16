package hashpass

import (
	"log"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return "", err
	}
	return hash, nil
}

func VerifyPassword(savedHash, passwordGuess string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(passwordGuess, savedHash)
	if err != nil {
		log.Printf("Failed to verify password: %v", err)
		return false, err
	}
	
	return match, nil
}