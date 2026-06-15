package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

type Balance struct {
	currency string
	amount   decimalx.Decimal
	frozen   decimalx.Decimal
}

func NewBalance(currency string, amount, frozen decimalx.Decimal) (Balance, error) {
	c, err := requireCurrency(currency)
	if err != nil {
		return Balance{}, err
	}
	if err := requireNonNegative("amount", amount); err != nil {
		return Balance{}, err
	}
	if err := requireNonNegative("frozen", frozen); err != nil {
		return Balance{}, err
	}
	return Balance{c, amount, frozen}, nil
}
func (b Balance) Currency() string                 { return b.currency }
func (b Balance) Amount() decimalx.Decimal         { return b.amount }
func (b Balance) Frozen() decimalx.Decimal         { return b.frozen }
func (b Balance) Total() (decimalx.Decimal, error) { return b.amount.Add(b.frozen), nil }

type balanceJSON struct {
	Currency string           `json:"currency"`
	Amount   decimalx.Decimal `json:"amount"`
	Frozen   decimalx.Decimal `json:"frozen"`
}

func (b Balance) MarshalJSON() ([]byte, error) {
	return json.Marshal(balanceJSON{b.currency, b.amount, b.frozen})
}
func (b *Balance) UnmarshalJSON(data []byte) error {
	var dto balanceJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	bal, err := NewBalance(dto.Currency, dto.Amount, dto.Frozen)
	if err != nil {
		return err
	}
	*b = bal
	return nil
}

type Portfolio struct {
	accountID   string
	balances    []Balance
	positions   []Position
	totalEquity decimalx.Decimal
	updatedAt   time.Time
}

func NewPortfolio(accountID string, balances []Balance, positions []Position) (Portfolio, error) {
	if err := requireID("account_id", accountID); err != nil {
		return Portfolio{}, err
	}
	total := decimalx.Zero()
	for _, b := range balances {
		bt, err := b.Total()
		if err != nil {
			return Portfolio{}, fmt.Errorf("domainx: balance total: %w", err)
		}
		total = total.Add(bt)
	}
	for _, p := range positions {
		mv, err := p.MarketValue(p.AveragePrice())
		if err != nil {
			return Portfolio{}, err
		}
		total = total.Add(mv)
	}
	return Portfolio{accountID: strings.TrimSpace(accountID), balances: append([]Balance(nil), balances...), positions: append([]Position(nil), positions...), totalEquity: total, updatedAt: time.Now().UTC()}, nil
}
func (p Portfolio) AccountID() string             { return p.accountID }
func (p Portfolio) Balances() []Balance           { return append([]Balance(nil), p.balances...) }
func (p Portfolio) Positions() []Position         { return append([]Position(nil), p.positions...) }
func (p Portfolio) TotalEquity() decimalx.Decimal { return p.totalEquity }
func (p Portfolio) UpdatedAt() time.Time          { return p.updatedAt }

type portfolioJSON struct {
	AccountID   string           `json:"account_id"`
	Balances    []Balance        `json:"balances"`
	Positions   []Position       `json:"positions"`
	TotalEquity decimalx.Decimal `json:"total_equity"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (p Portfolio) MarshalJSON() ([]byte, error) {
	return json.Marshal(portfolioJSON{p.accountID, p.Balances(), p.Positions(), p.totalEquity, p.updatedAt})
}
func (p *Portfolio) UnmarshalJSON(data []byte) error {
	var dto portfolioJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	pf, err := NewPortfolio(dto.AccountID, dto.Balances, dto.Positions)
	if err != nil {
		return err
	}
	pf.totalEquity = dto.TotalEquity
	pf.updatedAt = dto.UpdatedAt.UTC()
	*p = pf
	return nil
}
