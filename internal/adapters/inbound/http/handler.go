package http

import (
	"encoding/json"
	"errors"
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
	mux.HandleFunc("GET /promotions", h.listPromotions)
	mux.HandleFunc("GET /promotions/{id}", h.getPromotion)
	mux.HandleFunc("POST /promotions", h.createPromotion)
	mux.HandleFunc("PUT /promotions/{id}", h.updatePromotion)
	mux.HandleFunc("DELETE /promotions/{id}", h.deletePromotion)
	mux.HandleFunc("POST /promotions/apply", h.applyPromotion)
}

func (h *Handler) listPromotions(w http.ResponseWriter, _ *http.Request) {
	result, err := h.service.ListPromotions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getPromotion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	result, err := h.service.GetPromotion(id)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) createPromotion(w http.ResponseWriter, r *http.Request) {
	var req application.PromotionPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	result, err := h.service.CreatePromotion(req)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h *Handler) updatePromotion(w http.ResponseWriter, r *http.Request) {
	var req application.PromotionPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	result, err := h.service.UpdatePromotion(id, req)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deletePromotion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.DeletePromotion(id); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrPromotionNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, application.ErrPromotionExists):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
