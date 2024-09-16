package gettenders

import (
	"errors"
	"net/http"
	"strconv"
	"tender_service/internal/lib/response"
	models2 "tender_service/internal/storage/models"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	Limit      uint `validate:"gte=0"`
	OffSet     uint `validate:"gte=0"`
	SeviceType []string
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

type ResponseList struct {
	Response []Response
}

type TendersGetter interface {
	GetTenders(req Request) (ResponseList, error)
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

	serviceTypes := r.URL.Query()["service_type"]

	for _, el := range serviceTypes {
		e := models2.TenderServiceType(el)
		if !models2.ValidateTenderServiceType(e) {
			return errors.New("service type not correct")
		}
	}

	req.SeviceType = serviceTypes

	return nil

}

func New(ts TendersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		err := validateBadrequest(&req, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		res, err := ts.GetTenders(req)

		if err != nil {
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
