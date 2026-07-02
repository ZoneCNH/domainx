package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

type Fee struct {
	amount   decimalx.Decimal
	currency string
	feeType  FeeType
}

func NewFee(amount decimalx.Decimal, currency string, typ FeeType) (Fee, error) {
	if err := requireNonNegative("fee", amount); err != nil {
		return Fee{}, err
	}
	c, err := requireCurrency(currency)
	if err != nil {
		return Fee{}, err
	}
	if !typ.Valid() {
		return Fee{}, fmt.Errorf("domainx: unsupported fee type %q: %w", typ, ErrInvalidAmount)
	}
	return Fee{amount: amount, currency: c, feeType: typ}, nil
}

func (f Fee) Amount() decimalx.Decimal { return f.amount }
func (f Fee) Currency() string         { return f.currency }
func (f Fee) Type() FeeType            { return f.feeType }

type feeJSON struct {
	Amount   decimalx.Decimal `json:"amount"`
	Currency string           `json:"currency"`
	Type     FeeType          `json:"type"`
}

func (f Fee) MarshalJSON() ([]byte, error) {
	return json.Marshal(feeJSON{f.amount, f.currency, f.feeType})
}
func (f *Fee) UnmarshalJSON(data []byte) error {
	var dto feeJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	fee, err := NewFee(dto.Amount, dto.Currency, dto.Type)
	if err != nil {
		return err
	}
	*f = fee
	return nil
}

type Trade struct {
	id        string
	orderID   string
	symbol    string
	side      OrderSide
	quantity  decimalx.Decimal
	price     decimalx.Decimal
	fee       *Fee
	timestamp time.Time
}

type TradeOption func(*Trade)

func WithTradeID(id string) TradeOption      { return func(t *Trade) { t.id = strings.TrimSpace(id) } }
func WithTradeTime(ts time.Time) TradeOption { return func(t *Trade) { t.timestamp = ts.UTC() } }

func NewTrade(orderID, symbol string, side OrderSide, quantity, price decimalx.Decimal, fee *Fee, opts ...TradeOption) (Trade, error) {
	if err := requireID("order_id", orderID); err != nil {
		return Trade{}, err
	}
	s, err := requireSymbol(symbol)
	if err != nil {
		return Trade{}, err
	}
	if !side.Valid() {
		return Trade{}, fmt.Errorf("domainx: unsupported trade side %q: %w", side, ErrInvalidSide)
	}
	if err := requirePositive("quantity", quantity); err != nil {
		return Trade{}, err
	}
	if !price.IsPositive() {
		return Trade{}, fmt.Errorf("domainx: trade price must be positive: %w", ErrInvalidPrice)
	}
	tr := Trade{id: newID("trd"), orderID: strings.TrimSpace(orderID), symbol: s, side: side, quantity: quantity, price: price, timestamp: time.Now().UTC()}
	if fee != nil {
		copied := *fee
		tr.fee = &copied
	}
	for _, opt := range opts {
		opt(&tr)
	}
	if tr.id == "" {
		tr.id = newID("trd")
	}
	return tr, nil
}

func NewFill(orderID, symbol string, side OrderSide, quantity, price decimalx.Decimal, fee *Fee, opts ...TradeOption) (Fill, error) {
	return NewTrade(orderID, symbol, side, quantity, price, fee, opts...)
}

type Fill = Trade

func (t Trade) ID() string                 { return t.id }
func (t Trade) OrderID() string            { return t.orderID }
func (t Trade) Symbol() string             { return t.symbol }
func (t Trade) Side() OrderSide            { return t.side }
func (t Trade) Quantity() decimalx.Decimal { return t.quantity }
func (t Trade) Price() decimalx.Decimal    { return t.price }
func (t Trade) Fee() *Fee {
	if t.fee == nil {
		return nil
	}
	f := *t.fee
	return &f
}
func (t Trade) Timestamp() time.Time { return t.timestamp }

type tradeJSON struct {
	ID        string           `json:"id"`
	OrderID   string           `json:"order_id"`
	Symbol    string           `json:"symbol"`
	Side      OrderSide        `json:"side"`
	Quantity  decimalx.Decimal `json:"quantity"`
	Price     decimalx.Decimal `json:"price"`
	Fee       *Fee             `json:"fee,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

func (t Trade) MarshalJSON() ([]byte, error) {
	return json.Marshal(tradeJSON{t.id, t.orderID, t.symbol, t.side, t.quantity, t.price, t.Fee(), t.timestamp})
}
func (t *Trade) UnmarshalJSON(data []byte) error {
	var dto tradeJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	tr, err := NewTrade(dto.OrderID, dto.Symbol, dto.Side, dto.Quantity, dto.Price, dto.Fee, WithTradeID(dto.ID), WithTradeTime(dto.Timestamp))
	if err != nil {
		return err
	}
	*t = tr
	return nil
}
