package models

import "time"

type User struct {
	ID       int64
	Name     string
	Email    string
	HashPass string
}

type Session struct {
	ID             int64
	UserID         int64
	AccessToken    string
	RefreshToken   string
	AccessExpires  time.Time
	RefreshExpires time.Time
	Issued         time.Time
}

type AccountUsers struct {
	ID        int64
	UserID    int64
	AccountID int64
}

type Account struct {
	ID int64
}

type Course struct {
	ID        int64
	AccountID int64
}

type Section struct {
	ID       int64
	CourseID int64
}

type Block struct {
	ID        int64
	SectionID int64
}

type File struct {
	ID      int64
	BlockID int64
}
