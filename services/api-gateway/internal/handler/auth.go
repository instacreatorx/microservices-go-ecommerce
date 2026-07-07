package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type AuthHandler struct {
	userServiceURL string
	logger         *zap.Logger
}

func NewAuthHandler(userServiceURL string, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userServiceURL: userServiceURL,
		logger:         logger,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(h.userServiceURL+"/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		h.logger.Error("failed to forward register request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(resp.Body)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(h.userServiceURL+"/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		h.logger.Error("failed to forward login request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(resp.Body)
}
