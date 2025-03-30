package cms

import (
	"fmt"
	"net/http"

	"github.com/menaguilherme/trigon/configs"
	"github.com/menaguilherme/trigon/internal/core/store"
	"go.uber.org/zap"
)

type Handler struct {
	Logger  *zap.SugaredLogger
	Configs configs.Config
	Store   store.Storage
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "ok",
		"env":    h.Configs.Env,
	}

	ctx := r.Context()

	err := h.Store.Users.Create(ctx)
	if err != nil {
		fmt.Println("Hello")
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("CMS Health: " + data["env"]))
}
