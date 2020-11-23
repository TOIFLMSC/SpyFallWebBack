package model

import "github.com/dgrijalva/jwt-go"

// Token type
type Token struct {
	UserId uint
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
func (u *User) Sanitize() {
	u.Password = ""
}
