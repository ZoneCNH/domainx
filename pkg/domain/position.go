package domain

import (
	"encoding/json"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

type Position struct {
	symbol        string
	quantity      decimalx.Decimal
	avgPrice      decimalx.Decimal
	unrealizedPnL decimalx.Decimal
	realizedPnL   decimalx.Decimal
	updatedAt     time.Time
}

func NewPosition(symbol string, quantity, avgPrice decimalx.Decimal) (Position, error) {
	s, err := requireSymbol(symbol)
	if err != nil {
		return Position{}, err
	}
	if err := requireNonNegativePrice("average_price", avgPrice); err != nil {
		return Position{}, err
	}
	return Position{symbol: s, quantity: quantity, avgPrice: avgPrice, unrealizedPnL: decimalx.Zero(), realizedPnL: decimalx.Zero(), updatedAt: time.Now().UTC()}, nil
}

func (p Position) Symbol() string                       { return p.symbol }
func (p Position) Quantity() decimalx.Decimal           { return p.quantity }
func (p Position) AveragePrice() decimalx.Decimal       { return p.avgPrice }
func (p Position) UnrealizedPnLValue() decimalx.Decimal { return p.unrealizedPnL }
func (p Position) RealizedPnL() decimalx.Decimal        { return p.realizedPnL }
func (p Position) UpdatedAt() time.Time                 { return p.updatedAt }

func (p Position) WithQuantity(quantity, avgPrice decimalx.Decimal) (Position, error) {
	if err := requireNonNegativePrice("average_price", avgPrice); err != nil {
		return Position{}, err
	}
	copy := p
	copy.quantity = quantity
	copy.avgPrice = avgPrice
	copy.updatedAt = time.Now().UTC()
	return copy, nil
}

func (p Position) WithPnL(unrealized, realized decimalx.Decimal) Position {
	copy := p
	copy.unrealizedPnL = unrealized
	copy.realizedPnL = realized
	copy.updatedAt = time.Now().UTC()
	return copy
}

func (p Position) MarketValue(markPrice decimalx.Decimal) (decimalx.Decimal, error) {
	if err := requireNonNegativePrice("mark_price", markPrice); err != nil {
		return decimalx.Zero(), err
	}
	return p.quantity.Mul(markPrice)
}

func (p Position) UnrealizedPnL(markPrice decimalx.Decimal) (decimalx.Decimal, error) {
	if err := requireNonNegativePrice("mark_price", markPrice); err != nil {
		return decimalx.Zero(), err
	}
	diff, err := markPrice.Sub(p.avgPrice)
	if err != nil {
		return decimalx.Zero(), err
	}
	return diff.Mul(p.quantity)
}

type positionJSON struct {
	Symbol        string           `json:"symbol"`
	Quantity      decimalx.Decimal `json:"quantity"`
	AveragePrice  decimalx.Decimal `json:"average_price"`
	UnrealizedPnL decimalx.Decimal `json:"unrealized_pnl"`
	RealizedPnL   decimalx.Decimal `json:"realized_pnl"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func (p Position) MarshalJSON() ([]byte, error) {
	return json.Marshal(positionJSON{p.symbol, p.quantity, p.avgPrice, p.unrealizedPnL, p.realizedPnL, p.updatedAt})
}
func (p *Position) UnmarshalJSON(data []byte) error {
	var dto positionJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	pos, err := NewPosition(dto.Symbol, dto.Quantity, dto.AveragePrice)
	if err != nil {
		return err
	}
	pos.unrealizedPnL = dto.UnrealizedPnL
	pos.realizedPnL = dto.RealizedPnL
	pos.updatedAt = dto.UpdatedAt.UTC()
	*p = pos
	return nil
}
