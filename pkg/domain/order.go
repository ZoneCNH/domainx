package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

type Order struct {
	id        string
	symbol    string
	side      OrderSide
	orderType OrderType
	quantity  decimalx.Decimal
	price     decimalx.Decimal
	state     OrderState
	filledQty decimalx.Decimal
	avgPrice  decimalx.Decimal
	clientID  string
	createdAt time.Time
	updatedAt time.Time
}

type OrderOption func(*Order)

func WithOrderID(id string) OrderOption { return func(o *Order) { o.id = strings.TrimSpace(id) } }
func WithClientID(id string) OrderOption {
	return func(o *Order) { o.clientID = strings.TrimSpace(id) }
}
func WithOrderTime(t time.Time) OrderOption {
	return func(o *Order) { o.createdAt, o.updatedAt = t.UTC(), t.UTC() }
}
func WithOrderState(s OrderState) OrderOption { return func(o *Order) { o.state = s } }

func NewOrder(symbol string, side OrderSide, typ OrderType, quantity, price decimalx.Decimal, opts ...OrderOption) (Order, error) {
	s, err := requireSymbol(symbol)
	if err != nil {
		return Order{}, err
	}
	if !side.Valid() {
		return Order{}, fmt.Errorf("domainx: unsupported order side %q: %w", side, ErrInvalidSide)
	}
	if !typ.Valid() {
		return Order{}, fmt.Errorf("domainx: unsupported order type %q: %w", typ, ErrInvalidOrderType)
	}
	if err := requirePositive("quantity", quantity); err != nil {
		return Order{}, err
	}
	if typ == OrderTypeMarket {
		if err := requireNonNegativePrice("price", price); err != nil {
			return Order{}, err
		}
	} else if !price.IsPositive() {
		return Order{}, fmt.Errorf("domainx: price must be positive for %s orders: %w", typ, ErrInvalidPrice)
	}
	now := time.Now().UTC()
	o := Order{id: newID("ord"), symbol: s, side: side, orderType: typ, quantity: quantity, price: price, state: OrderStatePending, filledQty: decimalx.Zero(), avgPrice: decimalx.Zero(), createdAt: now, updatedAt: now}
	for _, opt := range opts {
		opt(&o)
	}
	if o.id == "" {
		o.id = newID("ord")
	}
	if !o.state.Valid() {
		return Order{}, fmt.Errorf("domainx: unsupported order state %q: %w", o.state, ErrInvalidState)
	}
	return o, nil
}

func (o Order) ID() string                       { return o.id }
func (o Order) Symbol() string                   { return o.symbol }
func (o Order) Side() OrderSide                  { return o.side }
func (o Order) Type() OrderType                  { return o.orderType }
func (o Order) Quantity() decimalx.Decimal       { return o.quantity }
func (o Order) Price() decimalx.Decimal          { return o.price }
func (o Order) State() OrderState                { return o.state }
func (o Order) FilledQuantity() decimalx.Decimal { return o.filledQty }
func (o Order) AveragePrice() decimalx.Decimal   { return o.avgPrice }
func (o Order) ClientID() string                 { return o.clientID }
func (o Order) CreatedAt() time.Time             { return o.createdAt }
func (o Order) UpdatedAt() time.Time             { return o.updatedAt }

func (o Order) TransitionTo(next OrderState) (Order, error) {
	if !next.Valid() {
		return Order{}, fmt.Errorf("domainx: unsupported order state %q: %w", next, ErrInvalidState)
	}
	if !o.state.CanTransitionTo(next) {
		return Order{}, fmt.Errorf("domainx: cannot transition order %s from %s to %s: %w", o.id, o.state, next, ErrInvalidTransition)
	}
	copy := o
	copy.state = next
	copy.updatedAt = time.Now().UTC()
	return copy, nil
}

func (o Order) WithFill(filledQty, avgPrice decimalx.Decimal) (Order, error) {
	if err := requireNonNegative("filled_quantity", filledQty); err != nil {
		return Order{}, err
	}
	if filledQty.GreaterThan(o.quantity) {
		return Order{}, fmt.Errorf("domainx: filled quantity exceeds order quantity: %w", ErrInvalidQuantity)
	}
	if err := requireNonNegativePrice("average_price", avgPrice); err != nil {
		return Order{}, err
	}
	copy := o
	copy.filledQty = filledQty
	copy.avgPrice = avgPrice
	copy.updatedAt = time.Now().UTC()
	return copy, nil
}

type orderJSON struct {
	ID             string           `json:"id"`
	Symbol         string           `json:"symbol"`
	Side           OrderSide        `json:"side"`
	Type           OrderType        `json:"type"`
	Quantity       decimalx.Decimal `json:"quantity"`
	Price          decimalx.Decimal `json:"price"`
	State          OrderState       `json:"state"`
	FilledQuantity decimalx.Decimal `json:"filled_quantity"`
	AveragePrice   decimalx.Decimal `json:"average_price"`
	ClientID       string           `json:"client_id,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(orderJSON{ID: o.id, Symbol: o.symbol, Side: o.side, Type: o.orderType, Quantity: o.quantity, Price: o.price, State: o.state, FilledQuantity: o.filledQty, AveragePrice: o.avgPrice, ClientID: o.clientID, CreatedAt: o.createdAt, UpdatedAt: o.updatedAt})
}

func (o *Order) UnmarshalJSON(data []byte) error {
	var dto orderJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	ord, err := NewOrder(dto.Symbol, dto.Side, dto.Type, dto.Quantity, dto.Price, WithOrderID(dto.ID), WithClientID(dto.ClientID), WithOrderTime(dto.CreatedAt), WithOrderState(dto.State))
	if err != nil {
		return err
	}
	ord.filledQty = dto.FilledQuantity
	ord.avgPrice = dto.AveragePrice
	ord.updatedAt = dto.UpdatedAt.UTC()
	*o = ord
	return nil
}
