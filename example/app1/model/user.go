package model

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

func (u *User) Authenticate(password string) bool {
	return u.Password == password
}
