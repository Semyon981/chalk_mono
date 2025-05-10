package http

import (
	"chalk/internal/transport/http/middleware"
	"chalk/internal/usecases"
	"net/http"
)

type Handler struct {
	mux    *http.ServeMux
	authUC usecases.AuthUseCase
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func NewHandler(authUC usecases.AuthUseCase) http.Handler {
	h := &Handler{
		authUC: authUC,
		mux:    http.NewServeMux(),
	}
	h.mux.HandleFunc("POST /auth/sendcode", h.SendCode)
	h.mux.HandleFunc("POST /auth/signup", h.SignUp)
	h.mux.HandleFunc("POST /auth/signin", h.SignIn)
	h.mux.HandleFunc("POST /auth/refresh", h.RefreshSession)

	/*
		c

	*/

	/*
		POST /accounts/create
		GET /accounts/myaccounts
		POST /accounts/getbyname
		POST /accounts/getbyid
		GET /accounts/:id/members
		POST /accounts/:id/members/add
		POST /accounts/:id/members/delete
		POST /accounts/:id/members/update
	*/

	/*
		GET /users/me
		GET /users/:id
		GET /users/:id/accounts
	*/

	/*
		/auth/sendcode
		/auth/signup
		/auth/signin
		/auth/refresh

		/accounts/getmyaccounts
		/accounts/create
		/accounts/getbyname
		/accounts/getbyid
	*/

	return middleware.LoggingMiddleware(h)
}
