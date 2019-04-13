package dataprovider

import (
	"context"

	"github.com/hummerd/gophercon/internal/model"
)

type NotificationStore interface {
	Insert(ctx context.Context, notification *model.Notification) error
}
