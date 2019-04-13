package controller

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hummerd/gophercon/internal/dataprovider"
	"github.com/hummerd/gophercon/internal/model"
	"github.com/hummerd/gophercon/internal/service"
)

const (
	sessionTypeToken = "token"
)

// NewApp creates an instance of App controller
func NewApp(
	sessionStore service.SessionStore,
	notificationStore dataprovider.NotificationStore,
) *App {
	h := App{
		sessionStore:      sessionStore,
		notificationStore: notificationStore,
	}

	return &h
}

type App struct {
	sessionStore      service.SessionStore
	notificationStore dataprovider.NotificationStore
}

func (ha *App) CreateNotification(ctx context.Context, notification *model.Notification) error {
	err := ha.notificationStore.Insert(ctx, notification)
	if err != nil {
		return errors.Wrapf(err, "creating notification %+v", notification)
	}

	return nil
}
