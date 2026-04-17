package main

import (
	"log"
	"time"

	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/erickgreco/dawg-patrol/pkg/db"
	"github.com/erickgreco/dawg-patrol/pkg/env"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
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

	// Wiring dependencies
	userstore := users.NewStore(dbpool)
	userservice := users.NewService(userstore)
	userhandler := users.NewHandler(userservice)

	app := &application{
		config: cfg,
		users:  userhandler,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
