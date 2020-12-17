package sqlstore

import (
	"database/sql"

	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/model"
	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/store"
	"github.com/lib/pq"
)

// LobbyRepository struct
type LobbyRepository struct {
	store *Store
}

// Create func
func (r *LobbyRepository) Create(l *model.Lobby) error {

	return r.store.db.QueryRow("INSERT INTO lobbies (token, locations, currentlocation, amountpl, amountspy, spyplayers, allplayers, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING token",
		l.Token,
		pq.Array(l.Locations),
		l.CurrentLocation,
		l.AmountPl,
		l.AmountSpy,
		pq.Array(l.SpyPlayers),
		pq.Array(l.AllPlayers),
		l.Status,
	).Scan(&l.Token)
}

// ConnectUserToLobby func
func (r *LobbyRepository) ConnectUserToLobby(l *model.Lobby) error {

	return r.store.db.QueryRow("UPDATE lobbies SET allplayers = $1 WHERE token = $2 RETURNING allplayers",
		pq.Array(l.AllPlayers),
		l.Token,
	).Scan(pq.Array(&l.AllPlayers))
}

// ChooseSpyPlayersInLobby func
func (r *LobbyRepository) ChooseSpyPlayersInLobby(l *model.Lobby) error {

	return r.store.db.QueryRow("UPDATE lobbies SET spyplayers = $1 WHERE token = $2 RETURNING spyplayers",
		pq.Array(l.SpyPlayers),
		l.Token,
	).Scan(pq.Array(&l.SpyPlayers))
}

// StartGame func
func (r *LobbyRepository) StartGame(l *model.Lobby) error {

	return r.store.db.QueryRow("UPDATE lobbies SET status = $1 WHERE token = $2 RETURNING status",
		"Started",
		l.Token,
	).Scan(&l.Token)
}

// WonForSpy func
func (r *LobbyRepository) WonForSpy(l *model.Lobby) (string, error) {

	return "Spy won", r.store.db.QueryRow("UPDATE lobbies SET status = $1 WHERE token = $2 RETURNING status",
		"Spy won",
		l.Token,
	).Scan(&l.Status)
}

// WonForPeaceful func
func (r *LobbyRepository) WonForPeaceful(l *model.Lobby) (string, error) {

	return "Peaceful won", r.store.db.QueryRow("UPDATE lobbies SET status = $1 WHERE token = $2 RETURNING status",
		"Peaceful won",
		l.Token,
	).Scan(&l.Status)
}

// FindByToken func
func (r *LobbyRepository) FindByToken(token string) (*model.Lobby, error) {
	l := &model.Lobby{}
	if err := r.store.db.QueryRow(
		"SELECT token, locations, currentlocation, amountpl, amountspy, spyplayers, allplayers, status FROM lobbies WHERE token = $1",
		token,
	).Scan(
		&l.Token,
		pq.Array(&l.Locations),
		&l.CurrentLocation,
		&l.AmountPl,
		&l.AmountSpy,
		pq.Array(&l.SpyPlayers),
		pq.Array(&l.AllPlayers),
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
