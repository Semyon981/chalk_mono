package repo

import "errors"

var (
	// abstract
	ErrUniqueViolation     = errors.New("unique violation")
	ErrRecordNotFound      = errors.New("record not found")
	ErrForeignKeyViolation = errors.New("foreign key violation")

	// specific
	ErrAccountNameAlreadyTaken = errors.New("account name already taken")
	ErrAccountOwnerNotFound    = errors.New("user with account_owner_id not found")
	ErrAccountNotFound         = errors.New("account not found")
	ErrUserNotFound            = errors.New("user not found")
	ErrAccountMemberNotFound   = errors.New("account member not found")
	ErrCourseNotFound          = errors.New("course not found")
	ErrModuleNotFound          = errors.New("module not found")
	ErrLessonNotFound          = errors.New("lesson not found")
	ErrBlockNotFound           = errors.New("block not found")
	ErrFileNotFound            = errors.New("file not found")

	ErrAlreadyEnrolled  = errors.New("user already enrolled in course")
	ErrNotEnrolled      = errors.New("user not enrolled in course")
	ErrUserNotInAccount = errors.New("user not found in specified account")
)
