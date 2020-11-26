package sqlstore

import (
	"database/sql"

	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/store"

	_ "github.com/lib/pq"
)

// Store struct
type Store struct {
	db              *sql.DB
	userRepository  *UserRepository
	lobbyRepository *LobbyRepository
}

// New func
func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// User func
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}

// Lobby func
func (s *Store) Lobby() store.LobbyRepository {
	if s.lobbyRepository != nil {
		return s.lobbyRepository
	}

	s.lobbyRepository = &LobbyRepository{
		store: s,
	}

	return s.lobbyRepository
}
