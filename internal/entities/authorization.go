package entities

type Authorization struct {
	SessionID    int64
	AccessToken  string
	RefreshToken string
}
