package middleware

import (
	"net/http"

	"github.com/rs/zerolog"
)

func Recover(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error().
						Interface("panic", err).
						Str("url", r.URL.String()).
						Str("method", r.Method).
						Msg("request panicked")
					panic(err)
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}
