package domain

import "fmt"

// MinAmountRule: tổng tiền phải >= threshold.
type MinAmountRule struct {
	Threshold float64
}

func (r MinAmountRule) IsSatisfiedBy(ctx OrderContext) bool {
	return ctx.Amount >= r.Threshold
}

func (r MinAmountRule) Description() string {
	return fmt.Sprintf("amount >= %.0f", r.Threshold)
}

// MinCardCountRule: số lượng thẻ phải >= threshold.
type MinCardCountRule struct {
	Threshold int
}

func (r MinCardCountRule) IsSatisfiedBy(ctx OrderContext) bool {
	return ctx.CardCount >= r.Threshold
}

func (r MinCardCountRule) Description() string {
	return fmt.Sprintf("card_count >= %d", r.Threshold)
}
