package middleware

import (
	"context"
	"net/http"

	"github.com/pborman/uuid"
)

type ridContextKey int

var (
	ridKey ridContextKey = 1
)

func GetRequestID(ctx context.Context) string {
	v, _ := ctx.Value(ridKey).(string)
	return v
}

func RequestID() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := ""

			xReqID := r.Header["X-Request-ID"]
			if len(xReqID) > 0 {
				reqID = xReqID[0]
			} else {
				reqID = uuid.New()
			}

			ctx := context.WithValue(r.Context(), ridKey, reqID)
			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}
