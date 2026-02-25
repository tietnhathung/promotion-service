package http

import (
	"encoding/json"
	"net/http"
	"promotion-service/internal/application"
)

type Handler struct {
	service *application.PromotionService
}

func NewHandler(service *application.PromotionService) *Handler {
	return &Handler{service: service}
}

type applyPromotionRequest struct {
	Amount    float64 `json:"amount"`
	CardCount int     `json:"card_count"`
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /promotions/apply", h.applyPromotion)
}

func (h *Handler) applyPromotion(w http.ResponseWriter, r *http.Request) {
	var req applyPromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	result, err := h.service.ApplyBestPromotion(application.ApplyPromotionInput{
		Amount:    req.Amount,
		CardCount: req.CardCount,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
