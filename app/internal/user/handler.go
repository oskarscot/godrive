package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpx "github.com/oskarscot/godrive/internal/platform"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Setup(mux *chi.Mux) {
	mux.Get("/user/me", h.userInfo)
}

func (h *Handler) userInfo(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, "Login Success")
}
