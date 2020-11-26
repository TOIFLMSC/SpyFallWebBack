package apiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

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
	s.router.HandleFunc("/user/new", s.usersCreateHandler()).Methods("POST")
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

func (s *server) usersCreateHandler() http.HandlerFunc {

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

		um := &model.User{
			Login:    req.Login,
			Password: req.Password,
		}

		if err := s.store.User().Create(um); err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		loc, err := s.store.User().FindByLogin(req.Login)
		if err != nil {
			u.Error(w, http.StatusUnprocessableEntity, err)
			return
		}

		if loc.ID <= 0 {
			u.Error(w, http.StatusNotFound, errors.New("Failed to create account"))
			return
		}

		tk := &model.Token{UserID: uint(loc.ID)}
		token := jwtg.NewWithClaims(jwtg.GetSigningMethod("HS256"), tk)
		tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
		loc.Token = tokenString

		loc.Sanitize()

		response := u.Message(true, "Account has been created")
		response["account"] = loc
		u.Respond(w, response)
	}
}
