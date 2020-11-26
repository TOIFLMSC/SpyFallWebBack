package model

import (
	u "github.com/TOIFLMSC/spyfall-web-backend/internal/app/utils"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Token type
type Token struct {
	UserID uint
	jwt.StandardClaims
}

// User type
type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Lobby    string `json:"lobby"`
}

// Sanitize func
func (user *User) Sanitize() {
	user.Password = ""
}

// Validate func
func (user *User) Validate() (map[string]interface{}, bool) {

	if len(user.Password) < 6 {
		return u.Message(false, "Password must be longer"), false
	}

	return u.Message(false, "Requirement passed"), true
}

// EncryptPassword func
func (user *User) EncryptPassword() (map[string]interface{}, bool) {

	if resp, ok := user.Validate(); !ok {
		return resp, ok
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	return u.Message(false, "Encrypted passed"), true
}

// ComparePassword func
func (user *User) ComparePassword(password string) (map[string]interface{}, bool) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(false, "Invalid password"), false
	}

	return u.Message(false, "Correct password"), true
}
