package container

import (
	"database/sql"
	"log"
)

type Container struct {
	db     *sql.DB
	logger *log.Logger
}

func NewContainer(db *sql.DB, logger *log.Logger) *Container {
	return &Container{db: db, logger: logger}
}
