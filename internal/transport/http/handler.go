package http

import (
	"chalk/internal/errors"
	"chalk/internal/transport/http/dto"
	"chalk/internal/transport/http/middleware"
	"chalk/internal/usecases"
	"chalk/pkg/log"
	"encoding/json"
	"io"
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

	return middleware.LoggingMiddleware(h)
}

func (h *Handler) SendCode(w http.ResponseWriter, r *http.Request) {
	req := dto.SendCodeRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}

	codeID, err := h.authUC.SendCodeToEmail(r.Context(), req.Email)
	if handleError(w, r, err) {
		return
	}

	resp := dto.SendCodeResponse{CodeID: codeID}
	encodeAndWriteJSONResponse(w, r, resp)
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	req := dto.SignUpRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}

	userID, err := h.authUC.SignUp(r.Context(),
		usecases.SignUpParams{
			Code:     req.Code,
			CodeID:   req.CodeID,
			Name:     req.Name,
			Password: req.Password,
		})
	if handleError(w, r, err) {
		return
	}

	resp := dto.SignUpResponse{UserID: userID}
	encodeAndWriteJSONResponse(w, r, resp)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	req := dto.SignInRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}

	a, err := h.authUC.SignIn(r.Context(), usecases.SignInParams{Email: req.Email, Password: req.Password})
	if handleError(w, r, err) {
		return
	}
	resp := dto.SignInResponse{SessionID: a.SessionID, AccessToken: a.AccessToken, RefreshToken: a.RefreshToken}
	encodeAndWriteJSONResponse(w, r, resp)
}

func (h *Handler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	req := dto.RefreshSessionRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}

	a, err := h.authUC.RefreshSession(r.Context(), usecases.RefreshSessionParams{RefreshToken: req.RefreshToken})
	if handleError(w, r, err) {
		return
	}

	resp := dto.RefreshSessionResponse{SessionID: a.SessionID, AccessToken: a.AccessToken, RefreshToken: a.RefreshToken}
	encodeAndWriteJSONResponse(w, r, resp)
}

func encodeAndWriteJSONResponse(w http.ResponseWriter, r *http.Request, src interface{}) {
	w.Header().Set("Content-Type", "application/json")
	responseData, err := json.Marshal(src)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to marshal response: " + err.Error()))
		return
	}
	_, err = w.Write(responseData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Warnf("failed to write response data: %v", err)
		return
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusInternalServerError)
		return false
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, dst)
	if err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func handleError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}

	if errors.IsUserError(err) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return true
	}

	log.Errorf("method: %s, error: %v", r.URL.Path, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal error"))
	return true
}
