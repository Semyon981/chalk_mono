package usecases

import (
	"chalk/internal/repo"
	"chalk/pkg/mailer"
	"context"
	crand "crypto/rand"
	"crypto/sha512"
	"fmt"
	"math/rand/v2"
	"net/mail"
	"strconv"
	"time"
)

type authUseCase struct {
	cr repo.AuthCodeRepo
	ur repo.UsersRepo

	passHashSalt string

	codeTTL time.Duration

	mailer mailer.Mailer

	smtpFromAddr string
	smtpFromName string
}

func (uc *authUseCase) SendCodeToEmail(ctx context.Context, address string) (string, error) {
	addr, err := mail.ParseAddress(address)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}
	code := strconv.Itoa(rand.IntN(1000000))
	if err := uc.sendMail(addr.String(), code); err != nil {
		return "", fmt.Errorf("failed to send mail: %w", err)
	}
	codeID := crand.Text()
	err = uc.cr.Set(ctx, codeID, repo.VerificationCode{Email: addr.String(), Code: code}, uc.codeTTL)
	if err != nil {
		return "", fmt.Errorf("failed to save code: %w", err)
	}
	return codeID, nil
}

type SignUpParams struct {
	CodeID   string
	Name     string
	Password string
}

func (uc *authUseCase) SignUp(ctx context.Context, params SignUpParams) (int64, error) {
	code, err := uc.cr.Get(ctx, params.CodeID)
	if err != nil {
		return 0, fmt.Errorf("failed to get code: %w", err)
	}

	pwd := sha512.New()
	pwd.Write([]byte(params.Password))
	passHashed := string(pwd.Sum([]byte(uc.passHashSalt)))

	userID, err := uc.ur.CreateUser(ctx, repo.CreateUserParams{
		Email:          code.Email,
		Name:           params.Name,
		PasswordHashed: passHashed,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, err
}

const messageTemplate = "Subject: Код подтверждения для аккаунта\nFrom: %s %s\nTo: %s\n\nВаш код - %s"

func (uc *authUseCase) sendMail(address string, code string) error {
	message := fmt.Appendf([]byte{}, messageTemplate, uc.smtpFromName, uc.smtpFromAddr, address, code)
	err := uc.mailer.SendMail(uc.smtpFromAddr, address, message)
	if err != nil {
		return fmt.Errorf("Failed to send mail: %w", err)
	}
	return nil
}
