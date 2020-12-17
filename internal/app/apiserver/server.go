package apiserver

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/jwt"
	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/model"
	"github.com/TOIFLMSC/spyfall-web-backend/internal/app/store"
	u "github.com/TOIFLMSC/spyfall-web-backend/internal/app/utils"
	jwtg "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type server struct {
	router *mux.Router
	logger *logrus.Logger
	store  store.Store
}

func newServer(store store.Store) *server {
	s := &server{
		router: mux.NewRouter(),
		logger: logrus.New(),
		store:  store,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"})))

	s.router.Use(jwt.JwtAuthentication)
	s.router.Use(s.logRequest)
	s.router.HandleFunc("/user/new", s.createUser()).Methods("POST")
	s.router.HandleFunc("/user/login", s.logUser()).Methods("POST")
	s.router.HandleFunc("/lobby/create", s.createLobby()).Methods("POST")
	s.router.HandleFunc("/lobby/adminconnect/{token}", s.connectLobby()).Methods("POST")
	s.router.HandleFunc("/lobby/connect/{token}", s.connectLobby()).Methods("POST")
	s.router.HandleFunc("/lobby/start/{token}", s.startGame()).Methods("POST")
	s.router.HandleFunc("/lobby/checkresult/{token}", s.checkResult()).Methods("GET")
	s.router.HandleFunc("/lobby/checklocation/{token}", s.checkLocation()).Methods("POST")
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
		})

		logger.Infof("started %s %s", r.Method, r.RequestURI)
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level

		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}

		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}

func (s *server) createUser() http.HandlerFunc {

	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			u.Error(w, http.StatusBadRequest, err)
			return
		}

		usermodel := &model.User{
			Login:    req.Login,
			Password: req.Password,
		}

		checkUser, err := s.store.User().FindByLogin(req.Login)
		if checkUser != nil {
			response := u.Message(false, "This username is already used. Please try another username or reauthorize")
			u.Respond(w, response)
			return
		}

		if err := s.store.User().Create(usermodel); err != nil {
			response := u.Message(false, "Error while creating user")
			u.Respond(w, response)
			return
		}

		locUser, err := s.store.User().FindByLogin(req.Login)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		if locUser.ID <= 0 {
			u.Error(w, http.StatusNotFound, errors.New("Failed to create account"))
			return
		}

		tk := &model.Token{UserID: uint(locUser.ID)}
		token := jwtg.NewWithClaims(jwtg.GetSigningMethod("HS256"), tk)
		tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
		locUser.Token = tokenString

		locUser.Sanitize()

		response := u.Message(true, "Account has been created")
		response["account"] = locUser
		u.Respond(w, response)
	}
}

func (s *server) logUser() http.HandlerFunc {

	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			u.Error(w, http.StatusBadRequest, err)
			return
		}

		usermodel := &model.User{
			Login:    req.Login,
			Password: req.Password,
		}

		loc, err := s.store.User().FindByLogin(req.Login)
		if err != nil && loc == nil {
			u.Error(w, http.StatusUnauthorized, err)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(loc.Password), []byte(usermodel.Password))
		if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
			response := u.Message(false, "Invalid login credentials. Please try again")
			u.Respond(w, response)
			return
		}

		usermodel.Sanitize()

		tk := &model.Token{UserID: uint(loc.ID)}
		token := jwtg.NewWithClaims(jwtg.GetSigningMethod("HS256"), tk)
		tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
		usermodel.Token = tokenString

		response := u.Message(true, "Logged in")
		response["account"] = usermodel
		u.Respond(w, response)
	}
}

func (s *server) createLobby() http.HandlerFunc {

	type request struct {
		AmountPl  int `json:"amountpl"`
		AmountSpy int `json:"amountspy"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			u.Error(w, http.StatusBadRequest, err)
			return
		}

		lobbymodel := &model.Lobby{
			AmountPl:  req.AmountPl,
			AmountSpy: req.AmountSpy,
		}

		token := u.TokenGenerator()
		for lobbymodel, err := s.store.Lobby().FindByToken(token); lobbymodel != nil && err == nil; {
			token = u.TokenGenerator()
		}

		lobbymodel.Token = token

		lobbymodel.Locations, lobbymodel.CurrentLocation = LocationsGenerator()

		lobbymodel.Status = "Created"

		if err := s.store.Lobby().Create(lobbymodel); err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		currentlobby, err := s.store.Lobby().FindByToken(lobbymodel.Token)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		currentlobby.CurrentLocation = ""

		response := u.Message(true, "Lobby has been created")
		response["lobby"] = currentlobby
		u.Respond(w, response)
	}
}

// connectLobby func
func (s *server) connectLobby() http.HandlerFunc {

	type request struct {
		Login string `json:"login"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		req := &request{}

		vars := mux.Vars(r)
		token := vars["token"]

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			u.Error(w, http.StatusBadRequest, err)
			return
		}

		currentlobby, err := s.store.Lobby().FindByToken(token)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		if status, err := s.store.Lobby().CheckStatus(token); status == "Started" && err == nil {
			response := u.Message(false, "Game has started already, you can't enter")
			u.Respond(w, response)
			return
		}

		currentlobby.AllPlayers = append(currentlobby.AllPlayers, req.Login)

		err = s.store.Lobby().ConnectUserToLobby(currentlobby)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		for {
			if status, err := s.store.Lobby().CheckStatus(token); status == "Started" && err == nil {

				connectedlobby, err := s.store.Lobby().FindByToken(token)
				if err != nil {
					u.Error(w, http.StatusUnprocessableEntity, err)
					return
				}

				if flag := u.Contains(connectedlobby.SpyPlayers, req.Login); flag == true {
					var cleararray []string
					connectedlobby.CurrentLocation = ""
					connectedlobby.SpyPlayers = cleararray
					response := u.Message(true, "Game has started, you are spy")
					response["lobby"] = connectedlobby
					u.Respond(w, response)
				} else {
					var cleararray []string
					connectedlobby.SpyPlayers = cleararray
					response := u.Message(true, "Game has started, you are peaceful")
					response["lobby"] = connectedlobby
					u.Respond(w, response)
				}
				break
			}
		}
	}
}

// checkLocation func
func (s *server) checkLocation() http.HandlerFunc {

	type checkrequest struct {
		Login    string `json:"login"`
		Location string `json:"location"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		token := vars["token"]

		cheklocreq := &checkrequest{}

		if err := json.NewDecoder(r.Body).Decode(cheklocreq); err != nil {
			u.Error(w, http.StatusBadRequest, err)
			return
		}

		connectedlobby, err := s.store.Lobby().FindByToken(token)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		if flag := u.Contains(connectedlobby.SpyPlayers, cheklocreq.Login); flag == true {
			if connectedlobby.CurrentLocation == cheklocreq.Location {
				result, err := s.store.Lobby().WonForSpy(connectedlobby)
				if err != nil {
					response := u.Message(false, "Unavailiable to end game for spy")
					u.Respond(w, response)
					return
				}
				response := u.Message(true, result)
				response["lobby"] = connectedlobby
				u.Respond(w, response)
			} else {
				result, err := s.store.Lobby().WonForPeaceful(connectedlobby)
				if err != nil {
					response := u.Message(false, "Unavailiable to end game for peaceful")
					u.Respond(w, response)
					return
				}
				response := u.Message(true, result)
				response["lobby"] = connectedlobby
				u.Respond(w, response)
			}
		} else {
			response := u.Message(false, "Misha, a ti krasava")
			u.Respond(w, response)
		}
		return
	}
}

func (s *server) checkResult() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		token := vars["token"]

		for {
			if status, err := s.store.Lobby().CheckStatus(token); status == "Spy won" && err == nil {

				connectedlobby, err := s.store.Lobby().FindByToken(token)
				if err != nil {
					u.Error(w, http.StatusUnprocessableEntity, err)
					return
				}

				response := u.Message(true, status)
				response["lobby"] = connectedlobby
				u.Respond(w, response)

				break
			} else if status, err := s.store.Lobby().CheckStatus(token); status == "Peaceful won" && err == nil {

				connectedlobby, err := s.store.Lobby().FindByToken(token)
				if err != nil {
					u.Error(w, http.StatusUnprocessableEntity, err)
					return
				}

				response := u.Message(true, status)
				response["lobby"] = connectedlobby
				u.Respond(w, response)

				break
			}
		}
	}
}

// startGame func
func (s *server) startGame() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		token := vars["token"]

		currentlobby, err := s.store.Lobby().FindByToken(token)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		playersarray := currentlobby.AllPlayers

		if len(playersarray) != currentlobby.AmountPl {
			response := u.Message(false, "Unable to start game, not enough players")
			u.Respond(w, response)
			return
		}

		var spyplayers []string = make([]string, 0, 20)
		j := currentlobby.AmountPl

		for i := currentlobby.AmountSpy; i > 0; i-- {
			a := rand.Intn(j)
			spyplayers = append(spyplayers, playersarray[a])
			playersarray = append(playersarray[:a], playersarray[a+1:]...)
			j--
		}

		currentlobby.SpyPlayers = spyplayers

		err = s.store.Lobby().ChooseSpyPlayersInLobby(currentlobby)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		err = s.store.Lobby().StartGame(currentlobby)
		if err != nil {
			response := u.Message(false, "Unable to start game")
			u.Respond(w, response)
			return
		}

		response := u.Message(true, "Game has started")
		u.Respond(w, response)
	}
}

// LocationsGenerator func
func LocationsGenerator() ([]string, string) {
	locations := []string{"Bank", "Hospital", "Military unit", "Casino",
		"Hollywood", "Titanic", "The Death Star", "Hotel",
		"Russian Railways", "Malibu Beach", "Police Station",
		"Restaurant", "University", "Lyceum", "SPA", "Plane"}
	rand.Seed(time.Now().UnixNano())
	locarray := locations
	var finallocarray []string = make([]string, 0, 20)
	for i := 12; i > 0; i-- {
		a := rand.Intn(i)
		finallocarray = append(finallocarray, locarray[a])
		locarray = append(locarray[:a], locarray[a+1:]...)
	}
	b := rand.Intn(12)
	return finallocarray, finallocarray[b]
}
