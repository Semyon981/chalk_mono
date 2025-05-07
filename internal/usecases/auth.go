package usecases

import (
	uerrors "chalk/internal/errors"
	"chalk/internal/repo"
	"chalk/internal/repo/models"
	"chalk/internal/usecases/entities"
	"chalk/pkg/log"
	"chalk/pkg/mailer"
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"errors"

	"fmt"
	"math/rand/v2"
	"net/mail"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	RefreshSession(ctx context.Context, params RefreshSessionParams) (entities.Authorization, error)
	SendCodeToEmail(ctx context.Context, address string) (string, error)
	SignIn(ctx context.Context, params SignInParams) (entities.Authorization, error)
	SignUp(ctx context.Context, params SignUpParams) (int64, error)
	ValidateToken(ctx context.Context, accessToken string) (entities.Session, error)
}

func NewAuthUseCase(
	cr repo.AuthCodeRepo,
	ur repo.UsersRepo,
	sr repo.SessionsRepo,
	codeTTL time.Duration,
	mailer mailer.Mailer,
	smtpFromAddr string,
	smtpFromName string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) AuthUseCase {
	return &authUseCase{
		cr:              cr,
		ur:              ur,
		sr:              sr,
		codeTTL:         codeTTL,
		mailer:          mailer,
		smtpFromAddr:    smtpFromAddr,
		smtpFromName:    smtpFromName,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

type authUseCase struct {
	cr repo.AuthCodeRepo
	ur repo.UsersRepo
	sr repo.SessionsRepo

	codeTTL time.Duration

	mailer mailer.Mailer

	smtpFromAddr string
	smtpFromName string

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func (uc *authUseCase) SendCodeToEmail(ctx context.Context, address string) (string, error) {
	_, err := mail.ParseAddress(address)
	if err != nil {
		return "", uerrors.ErrInvalidEmail
	}
	code := strconv.Itoa(rand.IntN(1000000))

	if err := uc.sendMail(address, code); err != nil {
		return "", fmt.Errorf("failed to send mail: %w", err)
	}
	codeID := crand.Text()
	err = uc.cr.Set(ctx, codeID, models.EmailCode{Email: address, Code: code}, uc.codeTTL)
	if err != nil {
		return "", fmt.Errorf("failed to save code: %w", err)
	}
	return codeID, nil
}

type SignUpParams struct {
	CodeID   string
	Code     string
	Name     string
	Password string
}

func (uc *authUseCase) SignUp(ctx context.Context, params SignUpParams) (int64, error) {
	code, err := uc.cr.Get(ctx, params.CodeID)
	if err != nil {
		if errors.Is(err, repo.ErrRecordNotFound) {
			return 0, uerrors.ErrInvalidCodeID
		}
		return 0, fmt.Errorf("failed to get code: %w", err)
	}

	if code.Code != params.Code {
		return 0, uerrors.ErrInvalidCode
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return 0, uerrors.ErrPasswordTooLong
		}
		return 0, err
	}

	userID, err := uc.ur.CreateUser(ctx, repo.CreateUserParams{
		Email:          code.Email,
		Name:           params.Name,
		PasswordHashed: string(hashedPassword),
	})
	if err != nil {
		if errors.Is(err, repo.ErrUniqueViolation) {
			return 0, uerrors.ErrUserAlreadyRegistered
		}
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	_, err = uc.cr.Delete(ctx, params.CodeID)
	if err != nil {
		log.Warnf("failed to delete codeID: %v", err)
	}

	return userID, err
}

type SignInParams struct {
	Email    string
	Password string
}

func (uc *authUseCase) SignIn(ctx context.Context, params SignInParams) (entities.Authorization, error) {
	user, err := uc.ur.GetUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, repo.ErrRecordNotFound) {
			return entities.Authorization{}, uerrors.ErrUserIsNotRegistered
		}
		return entities.Authorization{}, fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashPass), []byte(params.Password)); err != nil {
		return entities.Authorization{}, uerrors.ErrInvalidPassword
	}

	accessToken := genToken(32)
	refreshToken := genToken(32)

	sessionID, err := uc.sr.CreateSession(ctx, repo.CreateSessionParams{
		UserID:         user.ID,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		AccessExpires:  time.Now().Add(uc.accessTokenTTL),
		RefreshExpires: time.Now().Add(uc.refreshTokenTTL),
		Issued:         time.Now(),
	})
	if err != nil {
		return entities.Authorization{}, fmt.Errorf("create session failed: %w", err)
	}

	return entities.Authorization{SessionID: sessionID, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

type RefreshSessionParams struct {
	RefreshToken string
}

func (uc *authUseCase) RefreshSession(ctx context.Context, params RefreshSessionParams) (entities.Authorization, error) {
	newAccessToken := genToken(32)
	newRefreshToken := genToken(32)

	sessionID, err := uc.sr.UpdateSessionWithRefreshToken(ctx,
		repo.UpdateSessionWithRefreshTokenParams{
			RefreshToken:      params.RefreshToken,
			NewAccessToken:    newAccessToken,
			NewRefreshToken:   newRefreshToken,
			NewAccessExpires:  time.Now().Add(uc.accessTokenTTL),
			NewRefreshExpires: time.Now().Add(uc.accessTokenTTL),
		},
	)
	if err != nil {
		return entities.Authorization{}, uerrors.ErrInvalidRefreshToken
	}

	return entities.Authorization{
		SessionID:    sessionID,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func genToken(l int) string {
	data := make([]byte, l)
	crand.Read(data)
	return base64.RawURLEncoding.EncodeToString(data)
}

func (uc *authUseCase) ValidateToken(ctx context.Context, accessToken string) (entities.Session, error) {
	session, err := uc.sr.GetSessionByAccessToken(ctx, accessToken)
	if err != nil {
		if errors.Is(err, repo.ErrRecordNotFound) {
			return entities.Session{}, uerrors.ErrInvalidAccessToken
		}
	}
	if time.Now().After(session.AccessExpires) {
		return entities.Session{}, uerrors.ErrAccessTokenExpired
	}
	return entities.Session{
		ID:             session.ID,
		UserID:         session.UserID,
		AccessToken:    session.AccessToken,
		RefreshToken:   session.RefreshToken,
		AccessExpires:  session.AccessExpires,
		RefreshExpires: session.RefreshExpires,
		Issued:         session.Issued,
	}, nil
}

const messageTemplate = "Subject: Код подтверждения для аккаунта\nFrom: %s %s\nTo: %s\n\nВаш код - %s"

func (uc *authUseCase) sendMail(address string, code string) error {
	message := fmt.Appendf([]byte{}, messageTemplate, uc.smtpFromName, uc.smtpFromAddr, address, code)
	err := uc.mailer.SendMail(uc.smtpFromAddr, address, message)
	if err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}
	return nil
}
