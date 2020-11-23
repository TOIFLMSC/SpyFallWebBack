package model

// Lobby type
type Lobby struct {
	Token           string   `json:"token"`
	Locations       []string `json:"locations"`
	CurrentLocation string   `json:"currentloc"`
	AmountPl        int      `json:"amountpl"`
	AmountSpy       int      `json:"amountspy"`
	SpyPlayers      []string `json:"spyplayers"`
	AllPlayers      []string `json:"allplayers"`
	Status          string   `json:"status"`
}
