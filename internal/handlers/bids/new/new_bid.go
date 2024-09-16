package newbid

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"tender_service/internal/lib/response"
	models2 "tender_service/internal/storage/models"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Request struct {
	Name        string    `json:"name" validate:"required,max=100"`
	TenderId    uuid.UUID `json:"tenderId" validate:"required,uuid"`
	Description string    `json:"description" validate:"required,max=500"`
	AuthorType  string    `json:"authorType" validate:"required"`
	AuthorID    uuid.UUID `json:"authorId" validate:"required,uuid"`
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

type BidSaver interface {
	SaveBid(req Request) (Response, error)
}

func validateBadrequest(req *Request, r *http.Request) string {
	const op = "handlers.newBid.ValidateBadRequest"
	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		return "request body is empty"
	}
	if err != nil {
		slog.Info(err.Error(), slog.String("op", op))
		return "invalid request"
	}

	if !models2.ValidateBidAuthorType(models2.BidAuthorType(req.AuthorType)) {
		return "invalid authorType"
	}

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		errMsgs := response.ValidationError(validateErr)
		return errMsgs
	}

	return ""
}

func New(ts BidSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		if errMsg := validateBadrequest(&req, r); errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, response.Error(errMsg))
			return
		}

		res, err := ts.SaveBid(req)

		if err != nil {
			if errors.Is(err, response.ErrUserNotExists) || errors.Is(err, response.ErrTenderNotExists) {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			if errors.Is(err, response.ErrNoRights) {
				w.WriteHeader(http.StatusForbidden)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			if errors.Is(err, response.ErrTenderNotExists) {
				w.WriteHeader(http.StatusNotFound)
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
