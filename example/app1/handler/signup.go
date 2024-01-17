package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/yuemori/blueprinter/example/app1/model"
)

type SignupHandler struct {
	userRepository UserRepository
}

func NewSignupHandler(userRepository UserRepository) *SignupHandler {
	return &SignupHandler{userRepository: userRepository}
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	name := r.FormValue("name")

	user := &model.User{
		Email:    email,
		Password: password,
		Name:     name,
	}

	if err := h.userRepository.Save(context.Background(), user); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Hello, %s!", user.Name)
}
