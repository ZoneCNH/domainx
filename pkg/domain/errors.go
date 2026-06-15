package domain

import "errors"

var (
	ErrInvalidID         = errors.New("domainx: invalid_id")
	ErrInvalidSymbol     = errors.New("domainx: invalid_symbol")
	ErrInvalidSide       = errors.New("domainx: invalid_side")
	ErrInvalidOrderType  = errors.New("domainx: invalid_order_type")
	ErrInvalidState      = errors.New("domainx: invalid_order_state")
	ErrInvalidTransition = errors.New("domainx: invalid_transition")
	ErrInvalidQuantity   = errors.New("domainx: invalid_quantity")
	ErrInvalidPrice      = errors.New("domainx: invalid_price")
	ErrInvalidAmount     = errors.New("domainx: invalid_amount")
	ErrInvalidCurrency   = errors.New("domainx: invalid_currency")
	ErrInvalidAccount    = errors.New("domainx: invalid_account")
)
