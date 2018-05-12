package command

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/urfave/cli"
)

// HTTPCommand holds the necessary data for its execution
type HTTPCommand struct {
	BaseCommand
}

// NewHTTPCommand constructs HTTPCommand
func NewHTTPCommand(baseCommand BaseCommand) *HTTPCommand {
	return &HTTPCommand{baseCommand}
}

// Run ...
func (cmd *HTTPCommand) Run(cliCtx *cli.Context) error {
	_, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Init router
	router := cmd.newRouter(cliCtx.App.Version)

	return http.ListenAndServe(fmt.Sprintf(":%d", cmd.config.HTTP.Port), router)
}

func (cmd *HTTPCommand) newRouter(version string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(chiMiddleware.Recoverer)

	if cmd.config.Debug {
		r.Use(chiMiddleware.Logger)
	}

	// ROUTES
	// Endpoint shows the version of the api
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, map[string]string{"version": version})
	})

	return r
}
