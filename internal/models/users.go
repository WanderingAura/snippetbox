package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModelInterface interface {
	Get(id int) (*User, error)
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	UpdatePassword(id int, currentPassword, newPassword string) error
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m *UserModel) Get(id int) (*User, error) {
	stmt := `SELECT id, name, email, created, hashed_password FROM users WHERE id = $1`
	user := &User{}
	err := m.DB.QueryRow(context.Background(), stmt, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Created, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return user, nil
}

// Inserts inserts the given data into the user table
func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES($1, $2, $3, now())
	RETURNING id`
	_, err = m.DB.Exec(context.Background(), stmt, name, email, hashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && strings.Contains(pgErr.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) UpdatePassword(id int, currentPassword, newPassword string) error {
	user, err := m.Get(id)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt := `UPDATE users SET hashed_password=$1 WHERE id=$2`
	_, err = m.DB.Exec(context.Background(), stmt, string(newHashedPassword), id)

	return err
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM users WHERE email = $1`
	err := m.DB.QueryRow(context.Background(), stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT id FROM users WHERE id = $1)`

	err := m.DB.QueryRow(context.Background(), stmt, id).Scan(&exists)
	return exists, err
}
