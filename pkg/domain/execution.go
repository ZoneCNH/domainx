package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

type ExecutionReport struct {
	id                string
	orderID           string
	state             OrderState
	filledQuantity    decimalx.Decimal
	remainingQuantity decimalx.Decimal
	averagePrice      decimalx.Decimal
	lastPrice         decimalx.Decimal
	lastQuantity      decimalx.Decimal
	timestamp         time.Time
}

type ExecutionReportOption func(*ExecutionReport)

func WithExecutionReportID(id string) ExecutionReportOption {
	return func(r *ExecutionReport) { r.id = strings.TrimSpace(id) }
}
func WithExecutionReportTime(ts time.Time) ExecutionReportOption {
	return func(r *ExecutionReport) { r.timestamp = ts.UTC() }
}

func NewExecutionReport(orderID string, state OrderState, filled, remaining, avgPrice, lastPrice, lastQuantity decimalx.Decimal, opts ...ExecutionReportOption) (ExecutionReport, error) {
	if err := requireID("order_id", orderID); err != nil {
		return ExecutionReport{}, err
	}
	if !state.Valid() {
		return ExecutionReport{}, fmt.Errorf("domainx: unsupported order state %q: %w", state, ErrInvalidState)
	}
	for label, value := range map[string]decimalx.Decimal{"filled_quantity": filled, "remaining_quantity": remaining, "average_price": avgPrice, "last_price": lastPrice, "last_quantity": lastQuantity} {
		if err := requireNonNegative(label, value); err != nil {
			return ExecutionReport{}, err
		}
	}
	if state == OrderStateFilled && !remaining.IsZero() {
		return ExecutionReport{}, fmt.Errorf("domainx: filled execution report cannot have remaining quantity: %w", ErrInvalidQuantity)
	}
	r := ExecutionReport{id: newID("exec"), orderID: strings.TrimSpace(orderID), state: state, filledQuantity: filled, remainingQuantity: remaining, averagePrice: avgPrice, lastPrice: lastPrice, lastQuantity: lastQuantity, timestamp: time.Now().UTC()}
	for _, opt := range opts {
		opt(&r)
	}
	if r.id == "" {
		r.id = newID("exec")
	}
	return r, nil
}
func (r ExecutionReport) ID() string                          { return r.id }
func (r ExecutionReport) OrderID() string                     { return r.orderID }
func (r ExecutionReport) State() OrderState                   { return r.state }
func (r ExecutionReport) FilledQuantity() decimalx.Decimal    { return r.filledQuantity }
func (r ExecutionReport) RemainingQuantity() decimalx.Decimal { return r.remainingQuantity }
func (r ExecutionReport) AveragePrice() decimalx.Decimal      { return r.averagePrice }
func (r ExecutionReport) LastPrice() decimalx.Decimal         { return r.lastPrice }
func (r ExecutionReport) LastQuantity() decimalx.Decimal      { return r.lastQuantity }
func (r ExecutionReport) Timestamp() time.Time                { return r.timestamp }

type executionReportJSON struct {
	ID                string           `json:"id"`
	OrderID           string           `json:"order_id"`
	State             OrderState       `json:"state"`
	FilledQuantity    decimalx.Decimal `json:"filled_quantity"`
	RemainingQuantity decimalx.Decimal `json:"remaining_quantity"`
	AveragePrice      decimalx.Decimal `json:"average_price"`
	LastPrice         decimalx.Decimal `json:"last_price"`
	LastQuantity      decimalx.Decimal `json:"last_quantity"`
	Timestamp         time.Time        `json:"timestamp"`
}

func (r ExecutionReport) MarshalJSON() ([]byte, error) {
	return json.Marshal(executionReportJSON{r.id, r.orderID, r.state, r.filledQuantity, r.remainingQuantity, r.averagePrice, r.lastPrice, r.lastQuantity, r.timestamp})
}
func (r *ExecutionReport) UnmarshalJSON(data []byte) error {
	var dto executionReportJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	report, err := NewExecutionReport(dto.OrderID, dto.State, dto.FilledQuantity, dto.RemainingQuantity, dto.AveragePrice, dto.LastPrice, dto.LastQuantity, WithExecutionReportID(dto.ID), WithExecutionReportTime(dto.Timestamp))
	if err != nil {
		return err
	}
	*r = report
	return nil
}
