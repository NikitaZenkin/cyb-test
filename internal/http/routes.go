package http

import (
	"github.com/go-chi/chi/v5"
	swagger "github.com/swaggo/http-swagger/v2"

	_ "cyb-test/docs"
)

func (c *controller) mountRouter(r chi.Router, basePath string) {
	r.Get("/swagger/*", swagger.Handler())

	r.Route(basePath, func(r chi.Router) {
		r.Route("/fqdn", func(r chi.Router) {
			r.Post("/load", c.FQDNsLoad)
			r.Post("/list", c.FQDNsGet)
		})
	})
}
