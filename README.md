# Promotion Service (Golang + Hexagonal Architecture)

Service mẫu đơn giản để tính khuyến mãi theo rule:
- Số tiền tối thiểu (`MinAmountRule`)
- Số thẻ tối thiểu (`MinCardCountRule`)

Và hỗ trợ discount:
- Theo phần trăm (`PercentageDiscount`)
- Theo số tiền cố định (`FixedAmountDiscount`)

## Kiến trúc

```text
cmd/server                      # composition root
internal/domain                 # business rules (entities, rules, discount)
internal/application            # use case: apply promotion
internal/ports                  # interface output port
internal/adapters/outbound      # repository implementation (memory)
internal/adapters/inbound       # HTTP handler
```

## Chạy service

```bash
go run ./cmd/server
```

## API

`POST /promotions/apply`

Body:

```json
{
  "amount": 600000,
  "card_count": 2
}
```

Ví dụ response:

```json
{
  "promotion_id": "PROMO_PERCENT_10",
  "promotion_name": "Giảm 10% cho đơn từ 500k và có ít nhất 2 thẻ",
  "discount_amount": 60000,
  "final_amount": 540000,
  "discount_type_desc": "10%"
}
```

## Test

```bash
go test ./...
```
