package application

import (
	"fmt"
	"promotion-service/internal/domain"
	"promotion-service/internal/ports"
)

type PromotionService struct {
	repo ports.PromotionRepository
}

func NewPromotionService(repo ports.PromotionRepository) *PromotionService {
	return &PromotionService{repo: repo}
}

type ApplyPromotionInput struct {
	Amount    float64
	CardCount int
}

type ApplyPromotionResult struct {
	PromotionID      string  `json:"promotion_id,omitempty"`
	PromotionName    string  `json:"promotion_name,omitempty"`
	DiscountAmount   float64 `json:"discount_amount"`
	FinalAmount      float64 `json:"final_amount"`
	DiscountTypeDesc string  `json:"discount_type_desc,omitempty"`
}

// ApplyBestPromotion chọn promotion hợp lệ có mức giảm cao nhất.
func (s *PromotionService) ApplyBestPromotion(input ApplyPromotionInput) (ApplyPromotionResult, error) {
	if input.Amount <= 0 {
		return ApplyPromotionResult{}, fmt.Errorf("amount must be > 0")
	}

	promotions, err := s.repo.ListActive()
	if err != nil {
		return ApplyPromotionResult{}, err
	}

	ctx := domain.OrderContext{Amount: input.Amount, CardCount: input.CardCount}

	best := ApplyPromotionResult{FinalAmount: input.Amount}
	for _, promo := range promotions {
		if err := promo.Validate(); err != nil {
			continue
		}
		if !promo.IsApplicable(ctx) {
			continue
		}

		discount := promo.DiscountAmount(input.Amount)
		if discount > best.DiscountAmount {
			best = ApplyPromotionResult{
				PromotionID:      promo.ID,
				PromotionName:    promo.Name,
				DiscountAmount:   discount,
				FinalAmount:      input.Amount - discount,
				DiscountTypeDesc: promo.Discount.Description(),
			}
		}
	}

	return best, nil
}
