package main

import (
	"log"
	"time"

	"github.com/erickgreco/dawg-patrol/internal/auth"
	"github.com/erickgreco/dawg-patrol/internal/robots"
	"github.com/erickgreco/dawg-patrol/internal/users"
	"github.com/erickgreco/dawg-patrol/pkg/db"
	"github.com/erickgreco/dawg-patrol/pkg/env"
)

func main() {
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

	expiryStr := env.GetString("JWT_EXPIRY", "30m")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		log.Fatal("invalid JWT_EXPIRY")
	}

	userStore := users.NewUserStore(dbpool)
	tokenService := auth.NewTokenService(env.GetJWTKey("JWT_SECRET"), expiry)
	userService := users.NewUserService(userStore, tokenService)

	robotStore := robots.NewRobotsStore(dbpool)
	robotService := robots.NewRobotService(robotStore)

	Seed(userService, robotService)
	log.Println("seed completed")
}
