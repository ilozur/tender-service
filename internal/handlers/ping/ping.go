package ping

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
)

func New(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctx.Done():
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, "ok")
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
			render.JSON(w, r, "service unavailable, please try again")
		}
	}

}
