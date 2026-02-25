package domain

import "fmt"

// OrderContext chứa dữ liệu đầu vào để đánh giá khuyến mãi.
type OrderContext struct {
	Amount    float64
	CardCount int
}

// Rule là điều kiện áp dụng promotion.
type Rule interface {
	IsSatisfiedBy(ctx OrderContext) bool
	Description() string
}

// Discount biểu diễn cách tính giảm giá.
type Discount interface {
	Apply(amount float64) float64
	Description() string
}

// Promotion là aggregate root của domain khuyến mãi.
type Promotion struct {
	ID       string
	Name     string
	Rules    []Rule
	Discount Discount
}

func (p Promotion) IsApplicable(ctx OrderContext) bool {
	for _, rule := range p.Rules {
		if !rule.IsSatisfiedBy(ctx) {
			return false
		}
	}
	return true
}

func (p Promotion) DiscountAmount(amount float64) float64 {
	if p.Discount == nil {
		return 0
	}
	applied := p.Discount.Apply(amount)
	if applied < 0 {
		return 0
	}
	if applied > amount {
		return amount
	}
	return applied
}

func (p Promotion) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("promotion id is required")
	}
	if p.Name == "" {
		return fmt.Errorf("promotion name is required")
	}
	if p.Discount == nil {
		return fmt.Errorf("discount is required")
	}
	return nil
}
