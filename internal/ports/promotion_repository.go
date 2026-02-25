package ports

import "promotion-service/internal/domain"

// PromotionRepository là output port để lấy danh sách khuyến mãi.
type PromotionRepository interface {
	ListActive() ([]domain.Promotion, error)
}
