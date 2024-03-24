package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"cyb-test/internal/service"
)

type controller struct {
	server *http.Server
	router chi.Router
	logger *zap.Logger
	srv    *service.Service
}

func New(logger *zap.Logger, srv *service.Service, port, basePath string) (*controller, error) {
	r := NewRouter()

	cntrl := controller{
		logger: logger,
		router: r,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: r,
		},
		srv: srv,
	}

	cntrl.mountRouter(r, basePath)

	return &cntrl, nil
}

func (c *controller) ListenAndServe() error {
	return c.server.ListenAndServe()
}

func (c *controller) Close() error {
	return c.server.Close()
}
