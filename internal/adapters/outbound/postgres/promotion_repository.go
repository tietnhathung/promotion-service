package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"promotion-service/internal/domain"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultTimeout = 5 * time.Second

type PromotionRepository struct {
	db *sql.DB
}

type dbRule struct {
	Type      string  `json:"type"`
	Threshold float64 `json:"threshold"`
}

func NewPromotionRepository(databaseURL string) (*PromotionRepository, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	repo := &PromotionRepository{db: db}
	if err := repo.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

func (r *PromotionRepository) Close() error {
	if r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *PromotionRepository) ensureSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		CREATE TABLE IF NOT EXISTS promotions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			rules_json JSONB NOT NULL,
			discount_type TEXT NOT NULL,
			discount_value DOUBLE PRECISION NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *PromotionRepository) List() ([]domain.Promotion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
SELECT id, name, rules_json, discount_type, discount_value
FROM promotions
ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promotions []domain.Promotion
	for rows.Next() {
		promotion, scanErr := scanPromotion(rows.Scan)
		if scanErr != nil {
			return nil, scanErr
		}
		promotions = append(promotions, promotion)
	}
	return promotions, rows.Err()
}

func (r *PromotionRepository) GetByID(id string) (domain.Promotion, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	row := r.db.QueryRowContext(ctx, `
SELECT id, name, rules_json, discount_type, discount_value
FROM promotions
WHERE id = $1`, id)

	promotion, err := scanPromotion(row.Scan)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Promotion{}, false, nil
		}
		return domain.Promotion{}, false, err
	}
	return promotion, true, nil
}

func (r *PromotionRepository) Create(promotion domain.Promotion) error {
	rulesJSON, discountType, discountValue, err := serializePromotion(promotion)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err = r.db.ExecContext(ctx, `
INSERT INTO promotions (id, name, rules_json, discount_type, discount_value, is_active)
VALUES ($1, $2, $3, $4, $5, TRUE)`,
		promotion.ID, promotion.Name, rulesJSON, discountType, discountValue)
	return err
}

func (r *PromotionRepository) Update(id string, promotion domain.Promotion) (bool, error) {
	rulesJSON, discountType, discountValue, err := serializePromotion(promotion)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `
UPDATE promotions
SET name = $2,
	rules_json = $3,
	discount_type = $4,
	discount_value = $5,
	updated_at = NOW()
WHERE id = $1`,
		id, promotion.Name, rulesJSON, discountType, discountValue)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *PromotionRepository) Delete(id string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM promotions WHERE id = $1`, id)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *PromotionRepository) ListActive() ([]domain.Promotion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
SELECT id, name, rules_json, discount_type, discount_value
FROM promotions
WHERE is_active = TRUE
ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promotions []domain.Promotion
	for rows.Next() {
		promotion, scanErr := scanPromotion(rows.Scan)
		if scanErr != nil {
			return nil, scanErr
		}
		promotions = append(promotions, promotion)
	}
	return promotions, rows.Err()
}

type scannerFn func(dest ...any) error

func scanPromotion(scan scannerFn) (domain.Promotion, error) {
	var (
		id            string
		name          string
		rulesRaw      []byte
		discountType  string
		discountValue float64
	)

	if err := scan(&id, &name, &rulesRaw, &discountType, &discountValue); err != nil {
		return domain.Promotion{}, err
	}

	rules, err := parseRules(rulesRaw)
	if err != nil {
		return domain.Promotion{}, err
	}

	discount, err := parseDiscount(discountType, discountValue)
	if err != nil {
		return domain.Promotion{}, err
	}

	return domain.Promotion{
		ID:       id,
		Name:     name,
		Rules:    rules,
		Discount: discount,
	}, nil
}

func serializePromotion(promotion domain.Promotion) ([]byte, string, float64, error) {
	rules := make([]dbRule, 0, len(promotion.Rules))
	for _, rule := range promotion.Rules {
		switch v := rule.(type) {
		case domain.MinAmountRule:
			rules = append(rules, dbRule{Type: "min_amount", Threshold: v.Threshold})
		case domain.MinCardCountRule:
			rules = append(rules, dbRule{Type: "min_card_count", Threshold: float64(v.Threshold)})
		default:
			return nil, "", 0, fmt.Errorf("unsupported rule type in repository")
		}
	}

	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return nil, "", 0, err
	}

	switch v := promotion.Discount.(type) {
	case domain.PercentageDiscount:
		return rulesJSON, "percentage", v.Percent, nil
	case domain.FixedAmountDiscount:
		return rulesJSON, "fixed", v.Amount, nil
	default:
		return nil, "", 0, fmt.Errorf("unsupported discount type in repository")
	}
}

func parseRules(raw []byte) ([]domain.Rule, error) {
	var stored []dbRule
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, err
	}

	result := make([]domain.Rule, 0, len(stored))
	for _, rule := range stored {
		switch strings.ToLower(strings.TrimSpace(rule.Type)) {
		case "min_amount":
			result = append(result, domain.MinAmountRule{Threshold: rule.Threshold})
		case "min_card_count":
			result = append(result, domain.MinCardCountRule{Threshold: int(rule.Threshold)})
		default:
			return nil, fmt.Errorf("unsupported rule type: %s", rule.Type)
		}
	}
	return result, nil
}

func parseDiscount(kind string, value float64) (domain.Discount, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "percentage":
		return domain.PercentageDiscount{Percent: value}, nil
	case "fixed":
		return domain.FixedAmountDiscount{Amount: value}, nil
	default:
		return nil, fmt.Errorf("unsupported discount type: %s", kind)
	}
}
