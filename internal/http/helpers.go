package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

type Error struct {
	Error string `json:"error"`
}

type Values []string

func ResponseWithJSON(w http.ResponseWriter, log *zap.Logger, httpCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Error("response", zap.Error(err))
	}
}

func ResponseWithError(w http.ResponseWriter, log *zap.Logger, httpCode int, err error) {
	if err == nil {
		err = errors.New("empty error")
	}

	result := &Error{
		Error: err.Error(),
	}

	ResponseWithJSON(w, log, httpCode, result)
}

func NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 5))
	r.Use(
		cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{http.MethodGet, http.MethodPost},
			AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}).Handler)

	return r
}
