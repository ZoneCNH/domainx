package domain

import (
	"encoding/json"
	"fmt"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

type Exposure struct {
	accountID string
	gross     decimalx.Decimal
	net       decimalx.Decimal
	equity    decimalx.Decimal
}

func NewExposure(accountID string, gross, net, equity decimalx.Decimal) (Exposure, error) {
	if err := requireID("account_id", accountID); err != nil {
		return Exposure{}, err
	}
	if err := requireNonNegative("gross_exposure", gross); err != nil {
		return Exposure{}, err
	}
	if equity.IsNegative() {
		return Exposure{}, fmt.Errorf("domainx: equity must be non-negative: %w", ErrInvalidAmount)
	}
	return Exposure{accountID: accountID, gross: gross, net: net, equity: equity}, nil
}
func (e Exposure) AccountID() string        { return e.accountID }
func (e Exposure) Gross() decimalx.Decimal  { return e.gross }
func (e Exposure) Net() decimalx.Decimal    { return e.net }
func (e Exposure) Equity() decimalx.Decimal { return e.equity }
func (e Exposure) NetExposureRatio() (decimalx.Decimal, bool) {
	if e.equity.IsZero() {
		return decimalx.Zero(), false
	}
	r, err := e.net.Div(e.equity)
	if err != nil {
		return decimalx.Zero(), false
	}
	return r, true
}

type exposureJSON struct {
	AccountID string           `json:"account_id"`
	Gross     decimalx.Decimal `json:"gross"`
	Net       decimalx.Decimal `json:"net"`
	Equity    decimalx.Decimal `json:"equity"`
}

func (e Exposure) MarshalJSON() ([]byte, error) {
	return json.Marshal(exposureJSON{e.accountID, e.gross, e.net, e.equity})
}
func (e *Exposure) UnmarshalJSON(data []byte) error {
	var dto exposureJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	ex, err := NewExposure(dto.AccountID, dto.Gross, dto.Net, dto.Equity)
	if err != nil {
		return err
	}
	*e = ex
	return nil
}
