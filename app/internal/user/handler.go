package user

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oskarscot/godrive/internal/httpx"
	"github.com/oskarscot/godrive/internal/requestctx"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Setup(r chi.Router) {
	r.Get("/user/me", h.userInfo)
}

func (h *Handler) userInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := requestctx.UserIDFromContext(ctx)
	if !ok {
		httpx.Error(w, r, http.StatusUnauthorized, "unauthorized", "unable to retrieve session")
		return
	}

	user, err := h.service.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httpx.Error(w, r, http.StatusUnauthorized, "user not found", "no user found for the specified id")
			return
		}
		httpx.Error(w, r, http.StatusInternalServerError, "unauthorized", "unable to retrieve user details")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}
