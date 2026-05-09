package main

import (
	"log"
	"time"

	"github.com/erickgreco/dawg-patrol/internal/apimiddleware"
	"github.com/erickgreco/dawg-patrol/internal/auth"
	"github.com/erickgreco/dawg-patrol/internal/home"
	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/erickgreco/dawg-patrol/pkg/db"
	"github.com/erickgreco/dawg-patrol/pkg/env"
)

const version = "0.0.1"

func main() {

	expiryStr := env.GetString("JWT_EXPIRY", "30m")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		log.Fatal("invalid JWT_EXPIRY")
	}

	cfg := config{
		addr:      env.GetString("ADDR", ":8080"),
		jwtSecret: env.GetJWTKey("JWT_SECRET"),
		jwtExpiry: expiry,
	}

	dbcfg := db.DBConfig{
		MaxConns:              10,
		MinConns:              2,
		MaxConnLifeTime:       time.Hour,
		MaxConnIdleTime:       10 * time.Minute,
		HealthCheckPeriod:     time.Minute,
		MaxConnLifeTimeJitter: 5 * time.Minute,
	}

	dbpool, err := db.DBConn(env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5433/dawg-patrol?sslmode=disable"), dbcfg)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer dbpool.Close()
	log.Println("database connection established")

	// Wiring user dependencies
	userStore := users.NewUserStore(dbpool)
	tokenService := auth.NewTokenService(cfg.jwtSecret, cfg.jwtExpiry)
	userService := users.NewUserService(userStore, tokenService)
	userHandler := users.NewUserHandler(userService)

	// Wiring robot dependencies
	robotStore := robots.NewRobotsStore(dbpool)
	robotService := robots.NewRobotService(robotStore)
	robotHandler := robots.NewRobotHandler(robotService)

	middleware := apimiddleware.NewMiddleware(tokenService, robotService)

	// Wiring home dependencies
	homeService := home.NewHomeService(userService, robotService)
	homeHandler := home.NewHomeHandler(homeService)

	app := &application{
		config:     cfg,
		users:      userHandler,
		robots:     robotHandler,
		middleware: middleware,
		home:       homeHandler,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
