package http

import (
	"encoding/json"
	"net/http"

	"github.com/hummerd/gophercon/internal/config"
	"github.com/hummerd/gophercon/internal/model"
	"github.com/rs/zerolog"
)

func (srv *Server) getHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type versionResponse struct {
	Version string `json:"version"`
}

func (srv *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	respondOK(r.Context(), w, versionResponse{Version: config.Version})
}

type logLevel struct {
	Level string `json:"level" validate:"oneof=panic fatal error warning info debug"`
}

func (srv *Server) getLogLevel(w http.ResponseWriter, r *http.Request) {
	ll := zerolog.GlobalLevel().String()
	respondOK(r.Context(), w, logLevel{Level: ll})
}

func (srv *Server) setLogLevel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := new(logLevel)

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		respondError(ctx, w, err)
		return
	}

	lvl, err := zerolog.ParseLevel(request.Level)
	if err != nil {
		respondError(ctx, w, err)
		return
	}

	zerolog.SetGlobalLevel(lvl)

	respondOK(ctx, w, logLevel{Level: request.Level})
}

type createNotificationRequest struct {
	UserID *int64 `json:"user_id"`
	Title  string `json:"title" validate:"required"`
	Body   string `json:"body" validate:"required"`
	Type   string `json:"type" validate:"required"`
}

func (srv *Server) createNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := new(createNotificationRequest)

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		respondError(ctx, w, err)
		return
	}

	notification := &model.Notification{
		UserID: request.UserID,
		Title:  request.Title,
		Type:   request.Type,
		Body:   request.Body,
	}

	err := srv.app.CreateNotification(ctx, notification)
	if err != nil {
		respondError(ctx, w, err)
		return
	}

	respondOK(ctx, w, data{notification})
}
