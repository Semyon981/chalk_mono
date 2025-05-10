package middleware

import (
	"chalk/internal/errors"
	"chalk/internal/usecases"
	"chalk/pkg/log"
	"context"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler, auc usecases.AuthUseCase) http.Handler {
	return &authMiddleware{next: next, auc: auc}
}

const SessionKey string = "session"

type authMiddleware struct {
	next http.Handler
	auc  usecases.AuthUseCase
}

func (mw *authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		http.Error(w, errors.ErrInvalidAccessToken.Error(), http.StatusUnauthorized)
		return
	}
	session, err := mw.auc.ValidateToken(r.Context(), token)
	if err != nil {
		if errors.IsUserError(err) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			log.Errorf("validate token error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	r = r.WithContext(context.WithValue(r.Context(), SessionKey, session))
	mw.next.ServeHTTP(w, r)
}
