package sqlstore

import (
	"database/sql"
	"errors"

	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/model"
	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/store"
)

// UserRepository struct
type UserRepository struct {
	store *Store
}

// Create func
func (r *UserRepository) Create(u *model.User) error {

	if result, varbool := u.Validate(); result["status"] == false && result["message"] == "Password must be longer" && varbool == false {
		return errors.New("Password must be longer")
	}

	if _, varbool := u.EncryptPassword(); varbool == false {
		return errors.New("Unable to encrypt password")
	}

	return r.store.db.QueryRow("INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id",
		u.Login,
		u.Password,
	).Scan(&u.ID)
}

// Find func
func (r *UserRepository) Find(id int) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, login, password FROM users WHERE id = $1",
		id,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return u, nil
}

// FindByLogin func
func (r *UserRepository) FindByLogin(login string) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id, login, password FROM users WHERE login = $1",
		login,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return u, nil
}
