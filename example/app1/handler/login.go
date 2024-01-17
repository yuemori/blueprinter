package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/yuemori/blueprinter/example/app1/model"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Save(ctx context.Context, user *model.User) error
}

type LoginHandler struct {
	UserRepository UserRepository
}

func NewLoginHandler(userRepository UserRepository) *LoginHandler {
	return &LoginHandler{UserRepository: userRepository}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.UserRepository.FindByEmail(context.Background(), email)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user == nil || !user.Authenticate(password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "Hello, %s!", user.Name)
}
