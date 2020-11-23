package model

import "testing"

// TestUser func
func TestUser(t *testing.T) *User {
	return &User{
		Login:    "TestUserName",
		Password: "password",
	}
}

// TestLobby func
func TestLobby(t *testing.T) *Lobby {
	return &Lobby{
		Token:           "AAAAA",
		Locations:       []string{"TestLoc1", "TestLoc2", "TestLoc3", "TestLoc4"},
		CurrentLocation: "Test1",
		AmountPl:        5,
		AmountSpy:       1,
		SpyPlayers:      []string{"TestPlayer1"},
		AllPlayers:      []string{"TestPlayer1", "TestPlayer2", "TestPlayer3", "TestPlayer4", "TestPlayer5"},
	}
}
