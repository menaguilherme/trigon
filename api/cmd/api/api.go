package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/menaguilherme/trigon/configs"
	"github.com/menaguilherme/trigon/internal/auth"
	"github.com/menaguilherme/trigon/internal/store"
	"go.uber.org/zap"
)

type application struct {
	config        configs.Config
	logger        *zap.SugaredLogger
	store         store.Storage
	authenticator auth.Authenticator
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", app.RegisterUserHandler)
			r.Post("/login", app.LoginHandler)
			r.Post("/refresh", app.RefreshTokenHandler)
		})
	})

	return r

}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.Port,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server has started", "addr", app.config.Port, "env", app.config.Env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	fmt.Println(app.config.Port)

	app.logger.Infow("server has stopped", "addr", app.config.Port, "env", app.config.Env)

	return nil
}
