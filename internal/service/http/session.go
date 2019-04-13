package http

import (
	"context"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/hummerd/gophercon/internal/model"
)

var (
	// ErrSessionStoreUnavailble error is returned by package in case request can't be performed dut to unavailability.
	ErrSessionStoreUnavailble = errors.New("can't connect to session-store service")
)

// NewSessionStore creates new instance of the session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		client:     newCustomClient(withServicename("sessionstore")),
		sessionAPI: "example.com/api/v1/session/",
	}
}

// SessionStore implements dataprovider.SessionStore interface.
type SessionStore struct {
	logger *log.Logger
	client *httpClient

	sessionAPI string
}
type sessionByTokenResponse struct {
	Session *model.Session `json:"session"`
}

// GetSessionByToken gets user_id that is associated with token in session-store.
func (s *SessionStore) GetSessionByToken(ctx context.Context, token string) (*model.Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequest(http.MethodPost, s.sessionAPI+token, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var result sessionByTokenResponse
	err = s.client.DoJSON(ctx, req, &result)
	return result.Session, err
}
