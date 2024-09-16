package getbidstatus

import (
	"errors"
	"net/http"
	"tender_service/internal/lib/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	BidID    uuid.UUID `validate:"required,uuid"`
	UserName string
}

type Response struct {
	Status string
}

type BidStatusGetter interface {
	BidStatus(req Request) (Response, error)
}

func New(ts BidStatusGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		bidStr := chi.URLParam(r, "bidId")
		bidID, err := uuid.Parse(bidStr)

		if err != nil || bidStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid id"))
			return
		}
		req.BidID = bidID

		req.UserName = r.URL.Query().Get("username")
		res, err := ts.BidStatus(req)

		if err != nil {
			if errors.Is(err, response.ErrUserNotExists) {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			if errors.Is(err, response.ErrBidNotExists) {
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
