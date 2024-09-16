package patchtenderstatus

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"tender_service/internal/lib/response"
	models2 "tender_service/internal/storage/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Request struct {
	TenderID    uuid.UUID `json:"id"`
	UserName    string
	Name        string `json:"name" validate:"max=100"`
	Description string `json:"description" validate:"max=500"`
	ServiceType string `json:"serviceType"`
	Status      string `json:"status"`
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

type TenderStatusPatcher interface {
	PatchTender(req Request) (Response, error)
}

func validateBadrequest(req *Request, r *http.Request) string {
	err := render.DecodeJSON(r.Body, req)
	if errors.Is(err, io.EOF) {
		return "request body is empty"
	}
	if err != nil {
		slog.Info(err.Error())
		return "invalid request"
	}

	if !models2.ValidateTenderServiceType(models2.TenderServiceType(req.ServiceType)) && req.ServiceType != "" {
		return "incorrect service type"
	}

	if !models2.ValidateTenderStatus(models2.TenderStatus(req.Status)) && req.Status != "" {
		return "incorrect tender status"
	}

	if err := validator.New().Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)

		errMsgs := response.ValidationError(validateErr)
		return errMsgs
	}

	return ""
}

func New(ts TenderStatusPatcher) http.HandlerFunc {
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

		if errMsg := validateBadrequest(&req, r); errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, response.Error(errMsg))
			return
		}

		res, err := ts.PatchTender(req)

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
