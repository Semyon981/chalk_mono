package http

import (
	"chalk/internal/transport/http/dto"
	"chalk/internal/usecases"
	"net/http"
)

func (h *Handler) SendCode(w http.ResponseWriter, r *http.Request) {
	// s, ok := r.Context().Value(middleware.SessionKey).(entities.Session)
	// if !ok {
	// 	writeError(w, http.StatusInternalServerError, "method requires authentication")
	// }

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
