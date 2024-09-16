package new_tender

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
	Name            string    `json:"name" validate:"required,max=100"`
	Description     string    `json:"description" validate:"required,max=500"`
	ServiceType     string    `json:"serviceType" validate:"required"`
	OrganizationId  uuid.UUID `json:"organizationId" validate:"required,max=100,uuid"`
	CreatorUsername string    `json:"creatorUsername" validate:"required"`
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	Version     uint      `json:"version"`
	CreatedAt   string    `json:"createdAt"`
	Name        string    `json:"name" validate:"required,max=100"`
	Description string    `json:"description" validate:"required,max=500"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status" validate:"required"`
}

type TenderSaver interface {
	SaveTender(req Request) (Response, error)
}

func validateBadrequest(req *Request, r *http.Request) string {
	const op = "handlers.Tender.validateBadrequest"
	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		return "request body is empty"
	}

	if err != nil {
		slog.Info(err.Error(), slog.String("op", op))
		return "invalid request"
	}

	if !models2.ValidateTenderServiceType(models2.TenderServiceType(req.ServiceType)) {
		return "incorrect service type"
	}

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		errMsgs := response.ValidationError(validateErr)
		return errMsgs
	}

	return ""
}

func New(ts TenderSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req Request

		if errMsg := validateBadrequest(&req, r); errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, response.Error(errMsg))
			return
		}

		res, err := ts.SaveTender(req)

		if err != nil {
			if errors.Is(err, response.ErrUserNotExists) {
				w.WriteHeader(http.StatusUnauthorized)
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
