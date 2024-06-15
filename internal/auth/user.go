package auth

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID       string
	Email    string
	Password string
}

func NewUser(id, email, password string) *User {
	return &User{
		ID:       id,
		Email:    email,
		Password: password,
	}
}

func (u *User) CheckPasswordHash(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
