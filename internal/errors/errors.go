package errors

import (
	"encoding/json"
)

type userError struct {
	Code int    `json:"code"`
	Desc string `json:"description"`
}

func (e userError) Error() string {
	errMarshaled, _ := json.Marshal(e)
	return string(errMarshaled)
}

var (
	ErrInvalidAccessToken = userError{10, "invalid access token"}
	ErrAccessTokenExpired = userError{11, "access token expired"}

	ErrInvalidEmail          = userError{21, "invalid email"}
	ErrPasswordTooLong       = userError{22, "password too long"}
	ErrInvalidPassword       = userError{23, "invalid password"}
	ErrInvalidRefreshToken   = userError{24, "invalid or expired refresh token"}
	ErrInvalidCodeID         = userError{25, "invalid or expired code_id"}
	ErrInvalidCode           = userError{26, "invalid code"}
	ErrUserIsNotRegistered   = userError{27, "user is not registered"}
	ErrUserAlreadyRegistered = userError{28, "user already registered"}

	ErrUserNotFound = userError{31, "user not found"}

	ErrPermissionDenied     = userError{41, "you don't have permission to perform this action"}
	ErrUserAlreadyInAccount = userError{42, "user is already in the account"}
	ErrCannotRemoveYourself = userError{43, "you can't remove yourself"}
	ErrCannotChangeOwnRole  = userError{44, "you can't change your own role"}
	ErrCannotModifyOwner    = userError{45, "you can't modify or remove the owner"}
)

// var userErrors = map[error]struct{}{
// 	ErrInvalidAccessToken:  {},
// 	ErrAccessTokenExpired:  {},
// 	ErrInvalidEmail:        {},
// 	ErrPasswordTooLong:     {},
// 	ErrInvalidPassword:     {},
// 	ErrInvalidRefreshToken: {},
// }

func IsUserError(err error) bool {
	_, ok := err.(userError)
	return ok
}
