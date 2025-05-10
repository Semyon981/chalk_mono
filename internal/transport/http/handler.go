package http

import (
	"chalk/internal/transport/http/dto"
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

func (h *Handler) SendCode(w http.ResponseWriter, r *http.Request) {
	req := dto.SendCodeRequest{}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	codeID, err := h.authUC.SendCodeToEmail(r.Context(), req.Email)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.SendCodeResponse{CodeID: codeID})
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req dto.SignUpRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := h.authUC.SignUp(r.Context(), usecases.SignUpParams{
		Code:     req.Code,
		CodeID:   req.CodeID,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.SignUpResponse{UserID: userID})
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req dto.SignInRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.authUC.SignIn(r.Context(), usecases.SignInParams{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.SignInResponse{
		SessionID:    res.SessionID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

func (h *Handler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshSessionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.authUC.RefreshSession(r.Context(), usecases.RefreshSessionParams{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.RefreshSessionResponse{
		SessionID:    res.SessionID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}
