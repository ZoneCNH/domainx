package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

func (s OrderSide) Valid() bool { return s == OrderSideBuy || s == OrderSideSell }

type OrderType string

const (
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeStop      OrderType = "STOP"
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
)

func (t OrderType) Valid() bool {
	switch t {
	case OrderTypeMarket, OrderTypeLimit, OrderTypeStop, OrderTypeStopLimit:
		return true
	default:
		return false
	}
}

type OrderState string

const (
	OrderStatePending         OrderState = "PENDING"
	OrderStateSubmitted       OrderState = "SUBMITTED"
	OrderStatePartiallyFilled OrderState = "PARTIALLY_FILLED"
	OrderStateFilled          OrderState = "FILLED"
	OrderStateCancelled       OrderState = "CANCELLED"
	OrderStateRejected        OrderState = "REJECTED"
	OrderStateExpired         OrderState = "EXPIRED"
)

func (s OrderState) Valid() bool {
	switch s {
	case OrderStatePending, OrderStateSubmitted, OrderStatePartiallyFilled, OrderStateFilled, OrderStateCancelled, OrderStateRejected, OrderStateExpired:
		return true
	default:
		return false
	}
}

func (s OrderState) Terminal() bool {
	switch s {
	case OrderStateFilled, OrderStateCancelled, OrderStateRejected, OrderStateExpired:
		return true
	default:
		return false
	}
}

func (s OrderState) CanTransitionTo(next OrderState) bool {
	if !s.Valid() || !next.Valid() || s == next || s.Terminal() {
		return false
	}
	switch s {
	case OrderStatePending:
		return next == OrderStateSubmitted || next == OrderStateCancelled || next == OrderStateRejected || next == OrderStateExpired
	case OrderStateSubmitted:
		return next == OrderStatePartiallyFilled || next == OrderStateFilled || next == OrderStateCancelled || next == OrderStateRejected || next == OrderStateExpired
	case OrderStatePartiallyFilled:
		return next == OrderStateFilled || next == OrderStateCancelled || next == OrderStateRejected || next == OrderStateExpired
	default:
		return false
	}
}

type FeeType string

const (
	FeeTypeCommission FeeType = "COMMISSION"
	FeeTypeExchange   FeeType = "EXCHANGE"
	FeeTypeRegulatory FeeType = "REGULATORY"
	FeeTypeOther      FeeType = "OTHER"
)

func (t FeeType) Valid() bool {
	switch t {
	case FeeTypeCommission, FeeTypeExchange, FeeTypeRegulatory, FeeTypeOther:
		return true
	default:
		return false
	}
}

func normalizeSymbol(symbol string) string { return strings.ToUpper(strings.TrimSpace(symbol)) }
func normalizeCode(code string) string     { return strings.ToUpper(strings.TrimSpace(code)) }

func requireID(label, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("domainx: %s is required: %w", label, ErrInvalidID)
	}
	return nil
}

func requireSymbol(symbol string) (string, error) {
	s := normalizeSymbol(symbol)
	if s == "" {
		return "", fmt.Errorf("domainx: symbol is required: %w", ErrInvalidSymbol)
	}
	return s, nil
}

func requireCurrency(currency string) (string, error) {
	c := normalizeCode(currency)
	if c == "" {
		return "", fmt.Errorf("domainx: currency is required: %w", ErrInvalidCurrency)
	}
	return c, nil
}

func requirePositive(label string, value decimalx.Decimal) error {
	if !value.IsPositive() {
		return fmt.Errorf("domainx: %s must be positive: %w", label, ErrInvalidQuantity)
	}
	return nil
}

func requireNonNegative(label string, value decimalx.Decimal) error {
	if value.IsNegative() {
		return fmt.Errorf("domainx: %s must be non-negative: %w", label, ErrInvalidAmount)
	}
	return nil
}

func requireNonNegativePrice(label string, value decimalx.Decimal) error {
	if value.IsNegative() {
		return fmt.Errorf("domainx: %s must be non-negative: %w", label, ErrInvalidPrice)
	}
	return nil
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}
