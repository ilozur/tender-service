package getreviews

import (
	"errors"
	"net/http"
	"strconv"
	"tender_service/internal/lib/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	Limit             uint `validate:"gte=0"`
	OffSet            uint `validate:"gte=0"`
	AuthorUsername    string
	RequesterUsername string
	TenderID          uuid.UUID `validate:"uuid,required"`
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   string    `json:"createdAt"`
	Description string    `json:"description" validate:"max=500"`
}

type ResponseList struct {
	Response []Response
}

type ReviewsGetter interface {
	GetReviews(req Request) (ResponseList, error)
}

const (
	limitDefault  = 5
	offsetDefault = 0
)

func validateBadrequest(req *Request, r *http.Request) error {
	has := r.URL.Query().Has("limit")
	if has {
		limit := r.URL.Query().Get("limit")
		if value, err := strconv.Atoi(limit); err != nil {
			return err
		} else {
			if value < 0 {
				return errors.New("limit must be not negative")
			}
			req.Limit = uint(value)
		}
	} else {
		req.Limit = limitDefault
	}

	has = r.URL.Query().Has("offset")
	if has {
		offset := r.URL.Query().Get("offset")
		if value, err := strconv.Atoi(offset); err != nil {
			return err
		} else {
			if value < 0 {
				return errors.New("offset must be not negative")
			}
			req.OffSet = uint(value)
		}
	} else {
		req.OffSet = offsetDefault
	}

	has = r.URL.Query().Has("authorUsername")
	if has {
		req.AuthorUsername = r.URL.Query().Get("authorUsername")

	} else {
		return errors.New("authorUsername must be not empty")
	}

	has = r.URL.Query().Has("requesterUsername")
	if has {
		req.RequesterUsername = r.URL.Query().Get("requesterUsername")

	} else {
		return errors.New("requesterUsername must be not empty")
	}

	return nil
}

func New(ts ReviewsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req Request
		tenderStr := chi.URLParam(r, "tenderId")
		tenderID, err := uuid.Parse(tenderStr)

		if err != nil || tenderStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid  tender id"))
			return
		}
		req.TenderID = tenderID

		err = validateBadrequest(&req, r)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		res, err := ts.GetReviews(req)

		if err != nil {
			if errors.Is(err, response.ErrUserNotExists) {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			if errors.Is(err, response.ErrTenderNotExists) {
				w.WriteHeader(http.StatusNotFound)
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
		if res.Response == nil {
			res.Response = make([]Response, 0)
		}
		render.JSON(w, r, res.Response)

	}
}
