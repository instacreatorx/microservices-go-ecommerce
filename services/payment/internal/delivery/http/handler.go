package http

import (
	"encoding/json"
	"net/http"

	"github.com/ecommerce/payment/internal/domain"
	"github.com/ecommerce/payment/internal/usecase"
)

type PaymentHandler struct {
	uc *usecase.PaymentUsecase
}

func NewPaymentHandler(uc *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) Process(w http.ResponseWriter, r *http.Request) {
	var req domain.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payment, err := h.uc.Process(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, payment)
}

func (h *PaymentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	payment, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "payment not found")
		return
	}

	writeJSON(w, http.StatusOK, payment)
}

func (h *PaymentHandler) Refund(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	payment, err := h.uc.Refund(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, payment)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
