package repo

import "errors"

var (
	ErrUniqueViolation = errors.New("unique violation")
	ErrRecordNotFound  = errors.New("record not found")
)
