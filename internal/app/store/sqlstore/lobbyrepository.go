package sqlstore

import (
	"database/sql"

	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/model"
	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/store"
)

// LobbyRepository struct
type LobbyRepository struct {
	store *Store
}

// Create func
func (r *LobbyRepository) Create(l *model.Lobby) error {

	return r.store.db.QueryRow("INSERT INTO lobbies (token, locations, currentlocation, amountpl, amountspy, spyplayers, allplayers, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING token",
		l.Token,
		l.Locations,
		l.CurrentLocation,
		l.AmountPl,
		l.AmountSpy,
		l.SpyPlayers,
		l.AllPlayers,
		l.Status,
	).Scan(&l.Token)
}

// FindByToken func
func (r *LobbyRepository) FindByToken(token string) (*model.Lobby, error) {
	l := &model.Lobby{}
	if err := r.store.db.QueryRow(
		"SELECT token, locations, currentlocation, amountpl, amountspy, spyplayers, allplayers, status FROM lobbies WHERE token = $1",
		token,
	).Scan(
		&l.Token,
		&l.Locations,
		&l.CurrentLocation,
		&l.AmountPl,
		&l.AmountSpy,
		&l.SpyPlayers,
		&l.AllPlayers,
		&l.Status,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return l, nil
}

// CheckStatus func
func (r *LobbyRepository) CheckStatus(token string) (string, error) {
	l := &model.Lobby{}
	if err := r.store.db.QueryRow(
		"SELECT status FROM lobbies WHERE token = $1",
		token,
	).Scan(
		&l.Status,
	); err != nil {
		if err == sql.ErrNoRows {
			return "", store.ErrRecordNotFound
		}
		return "", err
	}

	return l.Status, nil
}
