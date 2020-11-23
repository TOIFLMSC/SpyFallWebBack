package apiserver

import (
"context"
"encoding/json"
"errors"
"net/http"
"time"

"github.com/TOIFLMSC/spyfall-web-backend/internal/app/model"
"github.com/TOIFLMSC/spyfall-web-backend/internal/app/jwt"
"github.com/TOIFLMSC/spyfall-web-backend/internal/app/store"
"github.com/gorilla/handlers"
"github.com/gorilla/mux"
"github.com/sirupsen/logrus"
)

type server struct {
	router	*mux.router
	logger	*logrus.logger
	store	store.Store
}

func newServer(store store.Store) *server {
	s := &server{
		router:	mux.NewRouter(),
		logger:	logrus.New(),
		store: store,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.Use(app.JwtAuthentication)
	s.router.Use(s.logRequest)
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
		})

		logger.InfoF("started %s %s", r.Method, r.RequestURI)
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

		logger.LogF(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start)
		)
	})
}