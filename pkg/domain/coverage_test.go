package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

func TestEnumBranchesAndOrderAccessors(t *testing.T) {
	if OrderType("BAD").Valid() || OrderState("BAD").Valid() || FeeType("BAD").Valid() {
		t.Fatal("bad enum should be invalid")
	}
	if !OrderStateFilled.Terminal() || OrderStateSubmitted.Terminal() {
		t.Fatal("terminal mismatch")
	}
	if !OrderStateSubmitted.CanTransitionTo(OrderStatePartiallyFilled) || !OrderStatePartiallyFilled.CanTransitionTo(OrderStateFilled) {
		t.Fatal("expected transitions")
	}
	if OrderStatePending.CanTransitionTo(OrderStatePending) || OrderState("BAD").CanTransitionTo(OrderStateFilled) {
		t.Fatal("unexpected transition")
	}
	o, err := NewOrder("btc", OrderSideSell, OrderTypeStop, dec("2"), dec("3"))
	if err != nil {
		t.Fatal(err)
	}
	filled, err := o.WithFill(dec("1"), dec("3"))
	if err != nil {
		t.Fatal(err)
	}
	if filled.Side() != OrderSideSell || filled.Type() != OrderTypeStop || !filled.Price().Equal(dec("3")) || !filled.FilledQuantity().Equal(dec("1")) || !filled.AveragePrice().Equal(dec("3")) || filled.CreatedAt().IsZero() || filled.UpdatedAt().IsZero() {
		t.Fatal("accessor mismatch")
	}
	if _, err := o.WithFill(dec("3"), dec("3")); !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("got %v", err)
	}
	if _, err := o.WithFill(dec("1"), dec("-1")); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewOrder("btc", OrderSideBuy, OrderType("BAD"), dec("1"), dec("1")); !errors.Is(err, ErrInvalidOrderType) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewOrder("btc", OrderSideBuy, OrderTypeMarket, dec("1"), dec("-1")); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewOrder("btc", OrderSideBuy, OrderTypeLimit, dec("1"), dec("1"), WithOrderState("BAD")); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("got %v", err)
	}
	if _, err := o.TransitionTo("BAD"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("got %v", err)
	}
}

func TestTradeFeeJSONAndAccessors(t *testing.T) {
	fee, err := NewFee(dec("0.1"), "usd", FeeTypeExchange)
	if err != nil {
		t.Fatal(err)
	}
	if !fee.Amount().Equal(dec("0.1")) || fee.Currency() != "USD" || fee.Type() != FeeTypeExchange {
		t.Fatal("fee accessors")
	}
	b, err := json.Marshal(fee)
	if err != nil {
		t.Fatal(err)
	}
	var fee2 Fee
	if err := json.Unmarshal(b, &fee2); err != nil {
		t.Fatal(err)
	}
	if !fee2.Amount().Equal(fee.Amount()) {
		t.Fatal("fee roundtrip")
	}
	if _, err := NewFee(dec("-1"), "USD", FeeTypeExchange); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewFee(decimalx.Zero(), "USD", FeeType("BAD")); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
	ts := time.Date(2026, 6, 15, 1, 0, 0, 0, time.UTC)
	tr, err := NewTrade("o1", "eth", OrderSideBuy, dec("2"), dec("5"), &fee, WithTradeID("t1"), WithTradeTime(ts))
	if err != nil {
		t.Fatal(err)
	}
	if tr.ID() != "t1" || tr.Side() != OrderSideBuy || !tr.Quantity().Equal(dec("2")) || !tr.Price().Equal(dec("5")) || !tr.Timestamp().Equal(ts) {
		t.Fatal("trade accessors")
	}
	b, err = json.Marshal(tr)
	if err != nil {
		t.Fatal(err)
	}
	var tr2 Trade
	if err := json.Unmarshal(b, &tr2); err != nil {
		t.Fatal(err)
	}
	if tr2.ID() != tr.ID() || tr2.Fee() == nil {
		t.Fatal("trade roundtrip")
	}
	if _, err := NewTrade("o1", "ETH", OrderSideBuy, dec("0"), dec("5"), nil); !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewTrade("o1", "ETH", OrderSideBuy, dec("1"), decimalx.Zero(), nil); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("got %v", err)
	}
}

func TestPositionJSONAndErrorBranches(t *testing.T) {
	if _, err := NewPosition("", dec("1"), dec("1")); !errors.Is(err, ErrInvalidSymbol) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewPosition("BTC", dec("1"), dec("-1")); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("got %v", err)
	}
	p, _ := NewPosition("btc", dec("1"), dec("10"))
	p = p.WithPnL(dec("2"), dec("3"))
	if p.Symbol() != "BTC" || !p.UnrealizedPnLValue().Equal(dec("2")) || !p.RealizedPnL().Equal(dec("3")) || p.UpdatedAt().IsZero() {
		t.Fatal("position accessors")
	}
	if _, err := p.WithQuantity(dec("1"), dec("-1")); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("got %v", err)
	}
	if _, err := p.MarketValue(dec("-1")); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("got %v", err)
	}
	if _, err := p.UnrealizedPnL(dec("-1")); !errors.Is(err, ErrInvalidPrice) {
		t.Fatalf("got %v", err)
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var p2 Position
	if err := json.Unmarshal(b, &p2); err != nil {
		t.Fatal(err)
	}
	if p2.Symbol() != p.Symbol() {
		t.Fatal("position roundtrip")
	}
}

func TestExposureJSONAndErrors(t *testing.T) {
	if _, err := NewExposure("", dec("1"), dec("1"), dec("1")); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewExposure("a", dec("-1"), dec("1"), dec("1")); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewExposure("a", dec("1"), dec("1"), dec("-1")); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
	e, _ := NewExposure("a", dec("2"), dec("1"), dec("4"))
	if e.AccountID() != "a" || !e.Gross().Equal(dec("2")) || !e.Net().Equal(dec("1")) || !e.Equity().Equal(dec("4")) {
		t.Fatal("exposure accessors")
	}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var e2 Exposure
	if err := json.Unmarshal(b, &e2); err != nil {
		t.Fatal(err)
	}
	if e2.AccountID() != e.AccountID() {
		t.Fatal("exposure roundtrip")
	}
}

func TestExecutionReportJSONAndAccessors(t *testing.T) {
	ts := time.Date(2026, 6, 15, 2, 0, 0, 0, time.UTC)
	r, err := NewExecutionReport("o1", OrderStateSubmitted, dec("1"), dec("2"), dec("3"), dec("4"), dec("5"), WithExecutionReportID("e1"), WithExecutionReportTime(ts))
	if err != nil {
		t.Fatal(err)
	}
	if r.ID() != "e1" || r.OrderID() != "o1" || r.State() != OrderStateSubmitted || !r.FilledQuantity().Equal(dec("1")) || !r.RemainingQuantity().Equal(dec("2")) || !r.AveragePrice().Equal(dec("3")) || !r.LastPrice().Equal(dec("4")) || !r.LastQuantity().Equal(dec("5")) || !r.Timestamp().Equal(ts) {
		t.Fatal("report accessors")
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var r2 ExecutionReport
	if err := json.Unmarshal(b, &r2); err != nil {
		t.Fatal(err)
	}
	if r2.ID() != r.ID() {
		t.Fatal("report roundtrip")
	}
	if _, err := NewExecutionReport("", OrderStateSubmitted, dec("1"), dec("0"), dec("1"), dec("1"), dec("1")); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewExecutionReport("o", OrderState("BAD"), dec("1"), dec("0"), dec("1"), dec("1"), dec("1")); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewExecutionReport("o", OrderStateSubmitted, dec("-1"), dec("0"), dec("1"), dec("1"), dec("1")); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
}

func TestBalancePortfolioJSONAndAccessors(t *testing.T) {
	if _, err := NewBalance("", dec("1"), decimalx.Zero()); !errors.Is(err, ErrInvalidCurrency) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewBalance("USD", dec("-1"), decimalx.Zero()); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewBalance("USD", dec("1"), dec("-1")); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v", err)
	}
	b, _ := NewBalance("usd", dec("1"), dec("2"))
	if b.Currency() != "USD" || !b.Amount().Equal(dec("1")) || !b.Frozen().Equal(dec("2")) {
		t.Fatal("balance accessors")
	}
	bb, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var b2 Balance
	if err := json.Unmarshal(bb, &b2); err != nil {
		t.Fatal(err)
	}
	p, _ := NewPosition("btc", dec("1"), dec("5"))
	pf, err := NewPortfolio("acct", []Balance{b}, []Position{p})
	if err != nil {
		t.Fatal(err)
	}
	if pf.AccountID() != "acct" || len(pf.Positions()) != 1 || pf.UpdatedAt().IsZero() {
		t.Fatal("portfolio accessors")
	}
	pb, err := json.Marshal(pf)
	if err != nil {
		t.Fatal(err)
	}
	var pf2 Portfolio
	if err := json.Unmarshal(pb, &pf2); err != nil {
		t.Fatal(err)
	}
	if pf2.AccountID() != pf.AccountID() {
		t.Fatal("portfolio roundtrip")
	}
	if _, err := NewPortfolio("", nil, nil); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("got %v", err)
	}
}
