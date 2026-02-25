package domain

import "fmt"

// PercentageDiscount giảm theo phần trăm.
type PercentageDiscount struct {
	Percent float64
}

func (d PercentageDiscount) Apply(amount float64) float64 {
	if d.Percent <= 0 {
		return 0
	}
	return amount * d.Percent / 100
}

func (d PercentageDiscount) Description() string {
	return fmt.Sprintf("%.0f%%", d.Percent)
}

// FixedAmountDiscount giảm theo số tiền cố định.
type FixedAmountDiscount struct {
	Amount float64
}

func (d FixedAmountDiscount) Apply(_ float64) float64 {
	if d.Amount <= 0 {
		return 0
	}
	return d.Amount
}

func (d FixedAmountDiscount) Description() string {
	return fmt.Sprintf("%.0f", d.Amount)
}
