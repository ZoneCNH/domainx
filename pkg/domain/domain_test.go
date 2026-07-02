package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ZoneCNH/decimalx/pkg/decimalx"
)

func dec(s string) decimalx.Decimal { return decimalx.MustFromString(s) }

func TestNewOrderValidatesAndNormalizes(t *testing.T) {
	o, err := NewOrder(" btc-usdt ", OrderSideBuy, OrderTypeLimit, dec("2"), dec("10.5"), WithOrderID("o1"), WithClientID("c1"))
	if err != nil {
		t.Fatal(err)
	}
	if o.Symbol() != "BTC-USDT" || o.ID() != "o1" || o.State() != OrderStatePending || o.ClientID() != "c1" {
		t.Fatalf("unexpected order: %+v", o)
	}
}

func TestNewOrderRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		name       string
		symbol     string
		side       OrderSide
		typ        OrderType
		qty, price decimalx.Decimal
		want       error
	}{
		{"symbol", " ", OrderSideBuy, OrderTypeLimit, dec("1"), dec("1"), ErrInvalidSymbol},
		{"side", "BTC", "BAD", OrderTypeLimit, dec("1"), dec("1"), ErrInvalidSide},
		{"qty", "BTC", OrderSideBuy, OrderTypeLimit, dec("0"), dec("1"), ErrInvalidQuantity},
		{"limit price", "BTC", OrderSideBuy, OrderTypeLimit, dec("1"), dec("0"), ErrInvalidPrice},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewOrder(tc.symbol, tc.side, tc.typ, tc.qty, tc.price)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestMarketOrderAllowsZeroPrice(t *testing.T) {
	if _, err := NewOrder("BTC", OrderSideBuy, OrderTypeMarket, dec("1"), decimalx.Zero()); err != nil {
		t.Fatal(err)
	}
}

func TestOrderTransitionIsImmutable(t *testing.T) {
	o, _ := NewOrder("BTC", OrderSideBuy, OrderTypeLimit, dec("1"), dec("2"), WithOrderID("o1"))
	next, err := o.TransitionTo(OrderStateSubmitted)
	if err != nil {
		t.Fatal(err)
	}
	if o.State() != OrderStatePending || next.State() != OrderStateSubmitted {
		t.Fatalf("not immutable transition")
	}
}
func TestOrderRejectsInvalidTransition(t *testing.T) {
	o, _ := NewOrder("BTC", OrderSideBuy, OrderTypeLimit, dec("1"), dec("2"))
	done, _ := o.TransitionTo(OrderStateCancelled)
	_, err := done.TransitionTo(OrderStateSubmitted)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("got %v", err)
	}
}

func TestOrderJSONRoundTripUsesSnakeCase(t *testing.T) {
	ts := time.Date(2026, 6, 15, 1, 2, 3, 0, time.UTC)
	o, _ := NewOrder("BTC", OrderSideBuy, OrderTypeLimit, dec("1.25"), dec("100"), WithOrderID("o1"), WithOrderTime(ts))
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "filled_quantity") || !strings.Contains(text, "\"quantity\":\"1.25\"") {
		t.Fatalf("bad json %s", text)
	}
	var got Order
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID() != o.ID() || !got.Quantity().Equal(o.Quantity()) {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestTradeAndFillValidation(t *testing.T) {
	fee, err := NewFee(dec("0.01"), "usd", FeeTypeCommission)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTrade("o1", "eth", OrderSideSell, dec("3"), dec("20"), &fee, WithTradeID("t1"))
	if err != nil {
		t.Fatal(err)
	}
	if tr.Fee() == &fee || tr.Symbol() != "ETH" {
		t.Fatalf("fee copied or symbol mismatch")
	}
	if _, err := NewTrade("", "ETH", OrderSideSell, dec("3"), dec("20"), nil); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("got %v", err)
	}
	f, err := NewFill("o1", "ETH", OrderSideSell, dec("1"), dec("20"), nil)
	if err != nil || f.OrderID() != "o1" {
		t.Fatalf("fill %v %v", f, err)
	}
}

func TestFeeRejectsCurrency(t *testing.T) {
	_, err := NewFee(dec("1"), "", FeeTypeCommission)
	if !errors.Is(err, ErrInvalidCurrency) {
		t.Fatalf("got %v", err)
	}
}

func TestPositionCalculationsAndImmutability(t *testing.T) {
	p, err := NewPosition("btc", dec("2"), dec("10"))
	if err != nil {
		t.Fatal(err)
	}
	updated, err := p.WithQuantity(dec("3"), dec("11"))
	if err != nil {
		t.Fatal(err)
	}
	if !p.Quantity().Equal(dec("2")) || !updated.Quantity().Equal(dec("3")) {
		t.Fatalf("mutation detected")
	}
	mv, err := updated.MarketValue(dec("12"))
	if err != nil || !mv.Equal(dec("36")) {
		t.Fatalf("mv %s err %v", mv, err)
	}
	pnl, err := updated.UnrealizedPnL(dec("12"))
	if err != nil || !pnl.Equal(dec("3")) {
		t.Fatalf("pnl %s err %v", pnl, err)
	}
}

func TestExposureRatioFailClosedOnZeroEquity(t *testing.T) {
	e, err := NewExposure("a1", dec("10"), dec("5"), decimalx.Zero())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := e.NetExposureRatio(); ok {
		t.Fatalf("expected ratio unavailable")
	}
	e, err = NewExposure("a1", dec("10"), dec("5"), dec("20"))
	if err != nil {
		t.Fatal(err)
	}
	r, ok := e.NetExposureRatio()
	if !ok || !r.Equal(dec("0.25")) {
		t.Fatalf("ratio %s ok %v", r, ok)
	}
}

func TestExecutionReportValidation(t *testing.T) {
	if _, err := NewExecutionReport("o1", OrderStateFilled, dec("1"), decimalx.Zero(), dec("10"), dec("10"), dec("1")); err != nil {
		t.Fatal(err)
	}
	_, err := NewExecutionReport("o1", OrderStateFilled, dec("1"), dec("1"), dec("10"), dec("10"), dec("1"))
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("got %v", err)
	}
}

func TestPortfolioTotalEquityAndCopies(t *testing.T) {
	b, _ := NewBalance("usd", dec("100"), dec("5"))
	p, _ := NewPosition("btc", dec("2"), dec("10"))
	pf, err := NewPortfolio("acct", []Balance{b}, []Position{p})
	if err != nil {
		t.Fatal(err)
	}
	if !pf.TotalEquity().Equal(dec("125")) {
		t.Fatalf("equity %s", pf.TotalEquity())
	}
	balances := pf.Balances()
	balances[0], _ = NewBalance("usd", dec("1"), decimalx.Zero())
	if !pf.TotalEquity().Equal(dec("125")) {
		t.Fatalf("external mutation affected portfolio")
	}
}

func TestConcurrentReadsAreSafe(t *testing.T) {
	o, _ := NewOrder("BTC", OrderSideBuy, OrderTypeLimit, dec("1"), dec("2"))
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = o.ID(); _ = o.Quantity().String(); _ = o.State() }()
	}
	wg.Wait()
}
