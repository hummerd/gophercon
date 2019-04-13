package model

import "time"

type Notification struct {
	ID       int        `json:"id"`
	UserID   *int64     `json:"-"`
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	Type     string     `json:"type"`
	FromTime *time.Time `json:"-"`
	TillTime *time.Time `json:"-"`
}
