package handler

import (
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type ProductHandler struct {
	productServiceURL string
	logger            *zap.Logger
}

func NewProductHandler(productServiceURL string, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productServiceURL: productServiceURL,
		logger:            logger,
	}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/v1/products?%s", h.productServiceURL, r.URL.RawQuery)
	resp, err := http.Get(url)
	if err != nil {
		h.logger.Error("failed to forward list products request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	url := fmt.Sprintf("%s/v1/products/%s", h.productServiceURL, id)
	resp, err := http.Get(url)
	if err != nil {
		h.logger.Error("failed to forward get product request", zap.Error(err))
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
