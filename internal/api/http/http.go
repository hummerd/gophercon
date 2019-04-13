package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	headerContentType    = "Content-Type"
	headerXRequestID     = "X-Request-ID"
	mimeApplicationJSON  = "application/json"
	mimeApplicationExcel = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

type data struct {
	Data interface{} `json:"data"`
}

type errResp struct {
	Error interface{} `json:"error"`
}

func respondError(ctx context.Context, w http.ResponseWriter, err error) {
	errCause := errors.Cause(err)
	respondJSON(ctx, w, http.StatusBadRequest, errResp{errCause})
}

func respondJSON(ctx context.Context, w http.ResponseWriter, code int, data interface{}) {
	if data == nil {
		w.WriteHeader(code)
		return
	}

	w.Header().Set(headerContentType, mimeApplicationJSON)
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(&data); err != nil {
		log.Error().Err(err).Msg("can write response")
	}
}

func respondRaw(ctx context.Context, w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func respondOK(ctx context.Context, w http.ResponseWriter, data interface{}) {
	respondJSON(ctx, w, http.StatusOK, data)
}

func respondNotFound(ctx context.Context, w http.ResponseWriter) {
	respondRaw(ctx, w, http.StatusNotFound)
}
