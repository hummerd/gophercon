package pg

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/hummerd/gophercon/internal/model"
)

func NewNotificationStore(db sqlx.ExtContext) *NotificationStore {
	return &NotificationStore{
		db: db,
	}
}

// NotificationStore is an notifications postgres store
type NotificationStore struct {
	db sqlx.ExtContext
}

// Insert inserts new notification
func (s *NotificationStore) Insert(ctx context.Context, notification *model.Notification) error {
	query, args, _ := sq.Insert("app.notifications").
		SetMap(map[string]interface{}{
			"type":      notification.Type,
			"title":     notification.Title,
			"body":      notification.Body,
			"user_id":   notification.UserID,
			"from_time": notification.FromTime,
			"till_time": notification.TillTime,
		}).
		Suffix("returning id;").
		PlaceholderFormat(sq.Dollar).ToSql()

	r := s.db.QueryRowxContext(ctx, query, args...)

	err := r.Scan(&notification.ID)
	if err != nil {
		return errors.Wrap(err, "can't scan notification id")
	}

	return nil
}

// GetByUser gets all global notifications or associated with user
func (s *NotificationStore) GetByUser(ctx context.Context, user *model.User) ([]*model.Notification, error) {
	notifications := make([]*model.Notification, 0)

	query, args, err := sq.Select("*").
		Where(
			sq.Or{
				sq.Eq{"user_id": user.ID},
				sq.Eq{"user_id": nil},
			},
		).
		From("app.notifications").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "creating sql query for getting user ids by user id")
	}

	err = sqlx.SelectContext(ctx, s.db, &notifications, query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "selecting user ids from database with query %s", query)
	}

	return notifications, nil
}
