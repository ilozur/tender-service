package tendersrollback

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
	TenderID uuid.UUID `json:"id"`
	UserName string    `validate:"required"`
	Version  uint      `validate:"required"`
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	Version     uint      `json:"version"`
	CreatedAt   string    `json:"createdAt"`
	Name        string    `json:"name" validate:"max=100"`
	Description string    `json:"description" validate:"max=500"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status"`
}

type TenderRollbacker interface {
	TenderRollback(req Request) (Response, error)
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

func New(ts TenderRollbacker) http.HandlerFunc {
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

		if errMsg := validateBadrequest(&req, r); errMsg != nil {
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, response.Error(errMsg.Error()))
			return
		}

		res, err := ts.TenderRollback(req)

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
		render.JSON(w, r, res)

	}
}
