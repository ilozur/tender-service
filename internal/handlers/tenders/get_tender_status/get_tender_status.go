package gettenderstatus

import (
	"errors"
	"net/http"
	"tender_service/internal/lib/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	TenderID uuid.UUID `validate:"required,uuid"`
	UserName string
}

type Response struct {
	Status string
}

type TenderStatusGetter interface {
	Status(req Request) (Response, error)
}

func New(ts TenderStatusGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		tenderStr := chi.URLParam(r, "tenderId")
		tenderID, err := uuid.Parse(tenderStr)

		if err != nil || tenderStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid id"))
			return
		}
		req.TenderID = tenderID

		req.UserName = r.URL.Query().Get("username")

		res, err := ts.Status(req)

		if err != nil {
			if errors.Is(err, response.ErrUserNotExists) {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			if errors.Is(err, response.ErrTenderNotExists) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			if errors.Is(err, response.ErrNoRights) {
				w.WriteHeader(http.StatusForbidden)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, res)

	}
}
