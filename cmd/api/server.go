package main

import (
	"log"
	"net/http"
	"time"

	"github.com/erickgreco/dawg-patrol/internal/apimiddleware"
	"github.com/erickgreco/dawg-patrol/internal/domain"
	"github.com/erickgreco/dawg-patrol/internal/home"
	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

type application struct {
	config     config
	users      *users.Handler
	robots     *robots.Handler
	middleware *apimiddleware.Middleware
	home       *home.HomeHandler
}

type config struct {
	addr      string
	jwtSecret string
	jwtExpiry time.Duration
}

// Method that mounts server router as well as register all server routes
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 60))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/register", func(r chi.Router) {
			r.Use(httprate.Limit(10, time.Minute, httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint)))
			r.Post("/", app.users.RegisterUserHandler)
		})

		r.Route("/login", func(r chi.Router) {
			r.Use(httprate.Limit(5, time.Minute, httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint)))
			r.Post("/", app.users.LogInHandler)
		})

		r.Route("/home", func(r chi.Router) {
			r.Use(app.middleware.AuthMiddleware)
			r.Use(httprate.Limit(100, time.Minute, httprate.WithKeyFuncs(app.middleware.KeyByUserID, httprate.KeyByEndpoint)))
			r.Get("/", app.home.HomePage)
		})

		r.Route("/profile", func(r chi.Router) {
			r.Use(app.middleware.AuthMiddleware)
			r.Use(httprate.Limit(100, time.Minute, httprate.WithKeyFuncs(app.middleware.KeyByUserID, httprate.KeyByEndpoint)))
			r.Get("/", app.users.UserProfileHandler)
			r.Post("/request-role-update", app.users.RequestRoleHandler)
		})

		r.Route("/robots", func(r chi.Router) {
			r.Use(app.middleware.AuthMiddleware)
			r.Use(httprate.Limit(100, time.Minute, httprate.WithKeyFuncs(app.middleware.KeyByUserID, httprate.KeyByEndpoint)))

			r.With(
				app.middleware.RequireRole(domain.RoleAdmin),
			).Post("/register-robot", app.robots.RegisterRobotHandler)

			r.With(
				app.middleware.RequireRole(domain.RoleAdmin, domain.RoleOperator),
			).Get("/idle-robots", app.robots.IdleRobotsHandler)

			r.With(
				app.middleware.RequireRole(domain.RoleAdmin, domain.RoleOperator),
			).Route("/{robotID}", func(r chi.Router) {
				r.Use(app.middleware.RobotContextMiddleware)

				r.Patch("/reserve-robot", app.home.ReserveRobot)
			})
		})
	})

	return r
}

// Method used to start server
func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Minute,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server iniciado en: %s", app.config.addr)

	return srv.ListenAndServe()
}
