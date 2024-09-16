package bidsubmitdecision

import (
	"errors"
	"net/http"
	"tender_service/internal/lib/response"
	models2 "tender_service/internal/storage/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Request struct {
	BidID    uuid.UUID `validate:"required,uuid"`
	UserName string
	Decision string `validate:"required"`
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

type BidDecisionSubmitter interface {
	BidSubmitDecision(req Request) (Response, error)
}

func validateBadrequest(req *Request) string {

	if !models2.ValidateBidDecision(models2.BidStatus(req.Decision)) {
		return "invalid bid decision"
	}

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		errMsgs := response.ValidationError(validateErr)
		return errMsgs
	}

	return ""
}

func New(ts BidDecisionSubmitter) http.HandlerFunc {
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
		req.Decision = r.URL.Query().Get("decision")

		if errMsg := validateBadrequest(&req); errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, response.Error(errMsg))
			return
		}

		res, err := ts.BidSubmitDecision(req)

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
