package dto

type SendCodeRequest struct {
	Email string `json:"email"`
}

type SendCodeResponse struct {
	CodeID string `json:"code_id"`
}

type SignUpRequest struct {
	CodeID   string `json:"code_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	UserID int64 `json:"user_id"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInResponse struct {
	SessionID    int64  `json:"session_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshSessionRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshSessionResponse struct {
	SessionID    int64  `json:"session_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
