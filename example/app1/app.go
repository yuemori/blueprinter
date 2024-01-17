//go:build !skip_blueprinter

package app1

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/yuemori/blueprinter/example/app1/container"
)

func Run() {
	dsn := os.Getenv("DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	container := container.NewContainer(db, logger)

	loginHandler := container.NewLoginHandler()
	signupHandler := container.NewSignupHandler()

	http.Handle("/login", loginHandler)
	http.Handle("/signup", signupHandler)

	http.ListenAndServe(":8080", nil)
}
