package repository

import (
	"context"
	"database/sql"

	"github.com/yuemori/blueprinter/example/app1/model"
)

type DB interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Logger interface {
	Printf(string, ...interface{})
}

type UserRepository struct {
	db     DB
	logger Logger
}

func NewUserRepository(db DB, logger Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	sql := "SELECT * FROM users WHERE email = ?"
	row := r.db.QueryRowContext(ctx, sql, email)

	r.logger.Printf("sql: %s", sql)

	user := &model.User{}
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Save(ctx context.Context, user *model.User) error {
	if user.ID == 0 {
		return r.insert(ctx, user)
	}
	return r.update(ctx, user)
}

func (r *UserRepository) insert(ctx context.Context, user *model.User) error {
	sql := "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"

	r.logger.Printf("sql: %s", sql)

	if _, err := r.db.ExecContext(ctx, sql, []interface{}{user.Name, user.Email, user.Password}...); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) update(ctx context.Context, user *model.User) error {
	sql := "UPDATE users SET name = ?, email = ?, password = ? WHERE id = ?"

	r.logger.Printf("sql: %s", sql)

	if _, err := r.db.ExecContext(ctx, sql, []interface{}{user.Name, user.Email, user.Password, user.ID}...); err != nil {
		return err
	}

	return nil
}
