package ports

import "promotion-service/internal/domain"

// PromotionRepository la output port quan ly promotion.
type PromotionRepository interface {
	List() ([]domain.Promotion, error)
	GetByID(id string) (domain.Promotion, bool, error)
	Create(promotion domain.Promotion) error
	Update(id string, promotion domain.Promotion) (bool, error)
	Delete(id string) (bool, error)
	ListActive() ([]domain.Promotion, error)
}
