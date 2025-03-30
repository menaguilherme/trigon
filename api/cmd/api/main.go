package main

import (
	"github.com/menaguilherme/trigon/configs"
	"github.com/menaguilherme/trigon/internal/core/db"
	"github.com/menaguilherme/trigon/internal/core/store"
	"go.uber.org/zap"
)

type application struct {
	config configs.Config
	logger *zap.SugaredLogger
	store  store.Storage
}

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

	store := store.NewStorage(db)

	app := &application{
		config: configs.Envs,
		logger: logger,
		store:  store,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
