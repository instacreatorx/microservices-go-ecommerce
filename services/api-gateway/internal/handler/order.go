package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type OrderHandler struct {
	orderServiceURL string
	logger          *zap.Logger
}

func NewOrderHandler(orderServiceURL string, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderServiceURL: orderServiceURL,
		logger:          logger,
	}
}

type CreateOrderRequest struct {
	Items            []OrderItem `json:"items"`
	ShippingAddress  string      `json:"shipping_address"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	proxyReq := map[string]interface{}{
		"user_id":          userID,
		"items":            req.Items,
		"shipping_address": req.ShippingAddress,
	}

	body, _ := json.Marshal(proxyReq)
	resp, err := http.Post(h.orderServiceURL+"/v1/orders", "application/json", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		h.logger.Error("failed to forward create order request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	url := fmt.Sprintf("%s/v1/orders?user_id=%s&%s", h.orderServiceURL, userID, r.URL.RawQuery)
	resp, err := http.Get(url)
	if err != nil {
		h.logger.Error("failed to forward list orders request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Context().Value("user_id").(string)
	url := fmt.Sprintf("%s/v1/orders/%s?user_id=%s", h.orderServiceURL, id, userID)
	resp, err := http.Get(url)
	if err != nil {
		h.logger.Error("failed to forward get order request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
