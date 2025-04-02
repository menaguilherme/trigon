package main

import (
	"github.com/menaguilherme/trigon/configs"
	"github.com/menaguilherme/trigon/internal/auth"
	"github.com/menaguilherme/trigon/internal/db"
	"github.com/menaguilherme/trigon/internal/store"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(
		configs.Envs.DB.ConnAddr,
		configs.Envs.DB.MaxOpenConns,
		configs.Envs.DB.MaxIdleConns,
		configs.Envs.DB.MaxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	jwtAuthenticator := auth.NewJWTAuthenticator(
		configs.Envs.Auth.Token.Secret,
		configs.Envs.Auth.Token.Aud,
		configs.Envs.Auth.Token.Iss,
	)

	store := store.NewStorage(db)

	app := &application{
		config:        configs.Envs,
		logger:        logger,
		store:         store,
		authenticator: jwtAuthenticator,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
