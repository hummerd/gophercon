package service

import (
	"context"

	"github.com/hummerd/gophercon/internal/model"
)

// SessionStore interface provides method to interacts with session-store.
type SessionStore interface {
	GetSessionByToken(ctx context.Context, token string) (*model.Session, error)
}
