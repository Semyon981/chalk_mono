package entities

import "time"

type Session struct {
	ID             int64
	UserID         int64
	AccessToken    string
	RefreshToken   string
	AccessExpires  time.Time
	RefreshExpires time.Time
	Issued         time.Time
}
