package getmybids

import (
	"errors"
	"net/http"
	"strconv"
	"tender_service/internal/lib/response"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	Limit    uint `validate:"gte=0"`
	OffSet   uint `validate:"gte=0"`
	UserName string
}

type Response struct {
	ID          uuid.UUID `json:"id"`
	Version     uint      `json:"version"`
	CreatedAt   string    `json:"createdAt"`
	Name        string    `json:"name" validate:"max=100"`
	AuthorType  string    `json:"authorType"`
	AuthorID    uuid.UUID `json:"authorId"`
	Description string    `json:"description" validate:"max=500"`
	Status      string    `json:"status"`
}

type ResponseList struct {
	Response []Response
}

type MyBidsGetter interface {
	GetMyBids(req Request) (ResponseList, error)
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

	has = r.URL.Query().Has("username")
	if has {
		req.UserName = r.URL.Query().Get("username")

	} else {
		return errors.New("username must be not empty")
	}

	return nil

}

func New(ts MyBidsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		err := validateBadrequest(&req, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		res, err := ts.GetMyBids(req)

		if err != nil {
			if errors.Is(err, response.ErrUserNotExists) {
				w.WriteHeader(http.StatusUnauthorized)
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
