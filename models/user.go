package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken = errors.New("models: email address is already in use")
)

type User struct {
	ID       int
	Email    string
	Password string
}

type UserService struct {
	DB *sql.DB
}

// store user's input into the database
// by doing so, create user
func (us *UserService) Create(email, password string) (*User, error) {
	email = strings.ToLower(email)
	// Hash the password
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %s", err)
	}
	passwordHash := string(passwordBytes)

	user := User{
		Email:    email,
		Password: passwordHash,
	}
	row := us.DB.QueryRow(`
	INSERT INTO users (email, password) VALUES
	($1, $2) RETURNING id;`, email, passwordHash)
	err = row.Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func (us *UserService) Authentificate(email, password string) (*User, error) {
	email = strings.ToLower(email)

	user := User{
		Email: email,
	}
	row := us.DB.QueryRow(`SELECT id, password FROM users WHERE email=$1;`, email)
	err := row.Scan(&user.ID, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("authentificate user: %s", err)
	}

	// compare the hash with provided password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authentificate user: %s", err)
	}

	return &user, nil
}
