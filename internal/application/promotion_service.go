package application

import (
	"errors"
	"fmt"
	"math"
	"promotion-service/internal/domain"
	"promotion-service/internal/ports"
	"strings"
)

var (
	ErrPromotionNotFound = errors.New("promotion not found")
	ErrPromotionExists   = errors.New("promotion already exists")
)

type PromotionService struct {
	repo ports.PromotionRepository
}

func NewPromotionService(repo ports.PromotionRepository) *PromotionService {
	return &PromotionService{repo: repo}
}

type RulePayload struct {
	Type      string  `json:"type"`
	Threshold float64 `json:"threshold"`
}

type DiscountPayload struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

type PromotionPayload struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Rules    []RulePayload   `json:"rules"`
	Discount DiscountPayload `json:"discount"`
}

func (s *PromotionService) ListPromotions() ([]PromotionPayload, error) {
	promotions, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	result := make([]PromotionPayload, 0, len(promotions))
	for _, promotion := range promotions {
		payload, convErr := toPromotionPayload(promotion)
		if convErr != nil {
			continue
		}
		result = append(result, payload)
	}
	return result, nil
}

func (s *PromotionService) GetPromotion(id string) (PromotionPayload, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return PromotionPayload{}, fmt.Errorf("promotion id is required")
	}

	promotion, found, err := s.repo.GetByID(id)
	if err != nil {
		return PromotionPayload{}, err
	}
	if !found {
		return PromotionPayload{}, ErrPromotionNotFound
	}
	return toPromotionPayload(promotion)
}

func (s *PromotionService) CreatePromotion(input PromotionPayload) (PromotionPayload, error) {
	promotion, err := fromPromotionPayload(input)
	if err != nil {
		return PromotionPayload{}, err
	}

	_, found, err := s.repo.GetByID(promotion.ID)
	if err != nil {
		return PromotionPayload{}, err
	}
	if found {
		return PromotionPayload{}, ErrPromotionExists
	}

	if err := s.repo.Create(promotion); err != nil {
		return PromotionPayload{}, err
	}
	return toPromotionPayload(promotion)
}

func (s *PromotionService) UpdatePromotion(id string, input PromotionPayload) (PromotionPayload, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return PromotionPayload{}, fmt.Errorf("promotion id is required")
	}

	input.ID = id
	promotion, err := fromPromotionPayload(input)
	if err != nil {
		return PromotionPayload{}, err
	}

	updated, err := s.repo.Update(id, promotion)
	if err != nil {
		return PromotionPayload{}, err
	}
	if !updated {
		return PromotionPayload{}, ErrPromotionNotFound
	}

	return toPromotionPayload(promotion)
}

func (s *PromotionService) DeletePromotion(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("promotion id is required")
	}

	deleted, err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrPromotionNotFound
	}
	return nil
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

// ApplyBestPromotion chon promotion hop le co muc giam cao nhat.
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

func fromPromotionPayload(input PromotionPayload) (domain.Promotion, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.Name = strings.TrimSpace(input.Name)

	promotion := domain.Promotion{
		ID:    input.ID,
		Name:  input.Name,
		Rules: make([]domain.Rule, 0, len(input.Rules)),
	}

	for _, ruleInput := range input.Rules {
		rule, err := buildRule(ruleInput)
		if err != nil {
			return domain.Promotion{}, err
		}
		promotion.Rules = append(promotion.Rules, rule)
	}

	discount, err := buildDiscount(input.Discount)
	if err != nil {
		return domain.Promotion{}, err
	}
	promotion.Discount = discount

	if err := promotion.Validate(); err != nil {
		return domain.Promotion{}, err
	}
	return promotion, nil
}

func toPromotionPayload(p domain.Promotion) (PromotionPayload, error) {
	payload := PromotionPayload{
		ID:    p.ID,
		Name:  p.Name,
		Rules: make([]RulePayload, 0, len(p.Rules)),
	}

	for _, rule := range p.Rules {
		rulePayload, err := toRulePayload(rule)
		if err != nil {
			return PromotionPayload{}, err
		}
		payload.Rules = append(payload.Rules, rulePayload)
	}

	discountPayload, err := toDiscountPayload(p.Discount)
	if err != nil {
		return PromotionPayload{}, err
	}
	payload.Discount = discountPayload
	return payload, nil
}

func buildRule(input RulePayload) (domain.Rule, error) {
	switch strings.ToLower(strings.TrimSpace(input.Type)) {
	case "min_amount":
		if input.Threshold <= 0 {
			return nil, fmt.Errorf("min_amount threshold must be > 0")
		}
		return domain.MinAmountRule{Threshold: input.Threshold}, nil
	case "min_card_count":
		if input.Threshold < 0 {
			return nil, fmt.Errorf("min_card_count threshold must be >= 0")
		}
		if !isWholeNumber(input.Threshold) {
			return nil, fmt.Errorf("min_card_count threshold must be an integer")
		}
		return domain.MinCardCountRule{Threshold: int(input.Threshold)}, nil
	default:
		return nil, fmt.Errorf("unsupported rule type: %s", input.Type)
	}
}

func buildDiscount(input DiscountPayload) (domain.Discount, error) {
	switch strings.ToLower(strings.TrimSpace(input.Type)) {
	case "percentage":
		if input.Value <= 0 {
			return nil, fmt.Errorf("percentage value must be > 0")
		}
		return domain.PercentageDiscount{Percent: input.Value}, nil
	case "fixed":
		if input.Value <= 0 {
			return nil, fmt.Errorf("fixed value must be > 0")
		}
		return domain.FixedAmountDiscount{Amount: input.Value}, nil
	default:
		return nil, fmt.Errorf("unsupported discount type: %s", input.Type)
	}
}

func toRulePayload(rule domain.Rule) (RulePayload, error) {
	switch v := rule.(type) {
	case domain.MinAmountRule:
		return RulePayload{Type: "min_amount", Threshold: v.Threshold}, nil
	case domain.MinCardCountRule:
		return RulePayload{Type: "min_card_count", Threshold: float64(v.Threshold)}, nil
	default:
		return RulePayload{}, fmt.Errorf("unsupported rule type in repository")
	}
}

func toDiscountPayload(discount domain.Discount) (DiscountPayload, error) {
	switch v := discount.(type) {
	case domain.PercentageDiscount:
		return DiscountPayload{Type: "percentage", Value: v.Percent}, nil
	case domain.FixedAmountDiscount:
		return DiscountPayload{Type: "fixed", Value: v.Amount}, nil
	default:
		return DiscountPayload{}, fmt.Errorf("unsupported discount type in repository")
	}
}

func isWholeNumber(value float64) bool {
	return math.Trunc(value) == value
}
