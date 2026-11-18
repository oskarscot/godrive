package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Setup(mux *chi.Mux) {
	mux.Get("/user/me", h.userInfo)
}

func (h *Handler) userInfo(w http.ResponseWriter, r *http.Request) {
}
