package bidsrollback

import (
	"errors"
	"net/http"
	"strconv"
	"tender_service/internal/lib/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Request struct {
	BidID    uuid.UUID `json:"id"`
	UserName string    `validate:"required"`
	Version  uint      `validate:"required"`
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	Version     uint      `json:"version"`
	CreatedAt   string    `json:"createdAt"`
	Name        string    `json:"name" validate:"required,max=100"`
	Description string    `json:"description" validate:"required,max=500"`
	AuthorType  string    `json:"authorType"`
	Status      string    `json:"status" validate:"required"`
	AuthorID    uuid.UUID `json:"authorId"`
}

type BidRollbacker interface {
	BidRollback(req Request) (Response, error)
}

func validateBadrequest(req *Request, r *http.Request) error {
	version := chi.URLParam(r, "version")
	if value, err := strconv.Atoi(version); err != nil {
		return err
	} else {
		if value < 0 {
			return errors.New("version must be not negative")
		}
		req.Version = uint(value)
	}

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		errMsgs := response.ValidationError(validateErr)
		return errors.New(errMsgs)
	}

	return nil
}

func New(ts BidRollbacker) http.HandlerFunc {
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

		if errMsg := validateBadrequest(&req, r); errMsg != nil {
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, response.Error(errMsg.Error()))
			return
		}

		res, err := ts.BidRollback(req)

		if err != nil {
			if errors.Is(err, response.ErrUserNotExists) {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			if errors.Is(err, response.ErrBidNotExists) {
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
		render.JSON(w, r, res)

	}
}
