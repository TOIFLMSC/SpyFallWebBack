package model

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
