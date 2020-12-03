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
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.Use(jwt.JwtAuthentication)
	s.router.Use(s.logRequest)
	s.router.HandleFunc("/user/new", s.createUser()).Methods("POST")
	s.router.HandleFunc("/user/login", s.logUser()).Methods("POST")
	s.router.HandleFunc("/lobby/create", s.createLobby()).Methods("POST")
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
			u.Error(w, http.StatusUnprocessableEntity, err)
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

		err = bcrypt.CompareHashAndPassword([]byte(req.Password), []byte(loc.Password))
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

		response := u.Message(true, "Lobby has been created")
		response["lobby"] = currentlobby
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
