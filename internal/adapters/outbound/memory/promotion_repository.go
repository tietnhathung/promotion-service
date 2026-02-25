package memory

import "promotion-service/internal/domain"

type PromotionRepository struct {
	promotions []domain.Promotion
}

func NewPromotionRepository() *PromotionRepository {
	return &PromotionRepository{
		promotions: []domain.Promotion{
			{
				ID:   "PROMO_PERCENT_10",
				Name: "Giảm 10% cho đơn từ 500k và có ít nhất 2 thẻ",
				Rules: []domain.Rule{
					domain.MinAmountRule{Threshold: 500_000},
					domain.MinCardCountRule{Threshold: 2},
				},
				Discount: domain.PercentageDiscount{Percent: 10},
			},
			{
				ID:   "PROMO_FIX_50K",
				Name: "Giảm cố định 50k cho đơn từ 300k",
				Rules: []domain.Rule{
					domain.MinAmountRule{Threshold: 300_000},
				},
				Discount: domain.FixedAmountDiscount{Amount: 50_000},
			},
		},
	}
}

func (r *PromotionRepository) ListActive() ([]domain.Promotion, error) {
	return r.promotions, nil
}
