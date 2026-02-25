package application

import (
	"testing"

	"promotion-service/internal/domain"
)

type stubRepo struct {
	promotions []domain.Promotion
}

func (s stubRepo) ListActive() ([]domain.Promotion, error) {
	return s.promotions, nil
}

func TestApplyBestPromotion_PicksHighestDiscount(t *testing.T) {
	svc := NewPromotionService(stubRepo{promotions: []domain.Promotion{
		{
			ID:   "percent",
			Name: "10%",
			Rules: []domain.Rule{
				domain.MinAmountRule{Threshold: 500_000},
			},
			Discount: domain.PercentageDiscount{Percent: 10},
		},
		{
			ID:       "fixed",
			Name:     "50k",
			Rules:    []domain.Rule{domain.MinAmountRule{Threshold: 300_000}},
			Discount: domain.FixedAmountDiscount{Amount: 50_000},
		},
	}})

	result, err := svc.ApplyBestPromotion(ApplyPromotionInput{Amount: 600_000, CardCount: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PromotionID != "percent" {
		t.Fatalf("expected percent, got %s", result.PromotionID)
	}
	if result.DiscountAmount != 60_000 {
		t.Fatalf("expected 60000, got %.0f", result.DiscountAmount)
	}
}

func TestApplyBestPromotion_NoPromotion(t *testing.T) {
	svc := NewPromotionService(stubRepo{promotions: []domain.Promotion{
		{
			ID:       "fixed",
			Name:     "50k",
			Rules:    []domain.Rule{domain.MinAmountRule{Threshold: 300_000}},
			Discount: domain.FixedAmountDiscount{Amount: 50_000},
		},
	}})

	result, err := svc.ApplyBestPromotion(ApplyPromotionInput{Amount: 100_000, CardCount: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DiscountAmount != 0 || result.FinalAmount != 100_000 {
		t.Fatalf("unexpected result: %+v", result)
	}
}
