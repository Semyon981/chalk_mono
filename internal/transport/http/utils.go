package http

import (
	"chalk/internal/errors"
	"chalk/pkg/log"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func readJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Warnf("write JSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeAppError(w http.ResponseWriter, err error) {
	if errors.IsUserError(err) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Errorf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
