package auth

import (
	"errors"
	"fmt"

	"avito-shop/internal/service/shop"
	"avito-shop/internal/service/shop/storage"

	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	storage storage.IStorage
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
func AuthenticateUser(s storage.IStorage, username, password, authKey string) (string, error) {
	storedPasswordHash, err := s.CheckAuth(username)

	if errors.Is(err, storage.ErrUserNotFound) {
		passwordHash, hashErr := HashPassword(password)
		if hashErr != nil {
			return "", fmt.Errorf("failed to hash password: %w", hashErr)
		}
		if addErr := s.AddNewUser(username, passwordHash); addErr != nil {
			return "", fmt.Errorf("failed to add new user: %w", addErr)
		}
		fmt.Println("User created:", username)
	} else if err != nil {
		return "", fmt.Errorf("failed to check authentication: %w", err)
	} else if checkErr := CheckPassword(storedPasswordHash, password); checkErr != nil {
		return "", fmt.Errorf("invalid password: %w", checkErr)
	}

	token, err := shop.GenerateJWT(authKey, username)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}
	return token, nil
}
