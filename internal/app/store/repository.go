package store

import "github.com/TOIFLMSC/spyfall-web-backend/internal/app/model"

// UserRepository interface
type UserRepository interface {
	Create(*model.User) error
	Find(int) (*model.User, error)
	FindByLogin(string) (*model.User, error)
}

// LobbyRepository interface
type LobbyRepository interface {
	Create(*model.Lobby) error
	FindByToken(string) (*model.Lobby, error)
	CheckStatus(string) (string, error)
	ConnectUserToLobby(*model.Lobby) error
	StartGame(*model.Lobby) error
	WonForSpy(*model.Lobby) (string, error)
	WonForPeaceful(*model.Lobby) (string, error)
	ChooseSpyPlayersInLobby(*model.Lobby) error
}
