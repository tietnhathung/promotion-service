package memory

import (
	"promotion-service/internal/domain"
	"sync"
)

type PromotionRepository struct {
	mu         sync.RWMutex
	promotions []domain.Promotion
}

func NewPromotionRepository() *PromotionRepository {
	return &PromotionRepository{
		promotions: []domain.Promotion{
			{
				ID:   "PROMO_PERCENT_10",
				Name: "Giam 10% cho don tu 500k va co it nhat 2 the",
				Rules: []domain.Rule{
					domain.MinAmountRule{Threshold: 500_000},
					domain.MinCardCountRule{Threshold: 2},
				},
				Discount: domain.PercentageDiscount{Percent: 10},
			},
			{
				ID:   "PROMO_FIX_50K",
				Name: "Giam co dinh 50k cho don tu 300k",
				Rules: []domain.Rule{
					domain.MinAmountRule{Threshold: 300_000},
				},
				Discount: domain.FixedAmountDiscount{Amount: 50_000},
			},
		},
	}
}

func (r *PromotionRepository) List() ([]domain.Promotion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Promotion, len(r.promotions))
	copy(result, r.promotions)
	return result, nil
}

func (r *PromotionRepository) GetByID(id string) (domain.Promotion, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.promotions {
		if p.ID == id {
			return p, true, nil
		}
	}
	return domain.Promotion{}, false, nil
}

func (r *PromotionRepository) Create(promotion domain.Promotion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.promotions = append(r.promotions, promotion)
	return nil
}

func (r *PromotionRepository) Update(id string, promotion domain.Promotion) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.promotions {
		if r.promotions[i].ID == id {
			r.promotions[i] = promotion
			return true, nil
		}
	}
	return false, nil
}

func (r *PromotionRepository) Delete(id string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.promotions {
		if r.promotions[i].ID == id {
			r.promotions = append(r.promotions[:i], r.promotions[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (r *PromotionRepository) ListActive() ([]domain.Promotion, error) {
	return r.List()
}
