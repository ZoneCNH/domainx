# domainx

`domainx` provides immutable L2.5 execution-domain value objects for ZoneCNH services.

## Public package

```go
import "github.com/ZoneCNH/domainx/pkg/domain"
```

## v1.0.0 value objects

- `Order` — validated side/type/state/quantity/price with explicit state transitions.
- `Trade` / `Fill` — immutable execution fill records with optional copied `Fee`.
- `Position` — quantity, average price, market value, and unrealized PnL helpers.
- `Exposure` — gross/net/equity exposure with fail-closed ratio calculation.
- `ExecutionReport` — broker/exchange execution-state snapshot validation.
- `Portfolio` / `Balance` — copied holdings and total-equity snapshot.

All money/quantity fields use `github.com/ZoneCNH/decimalx/pkg/decimalx.Decimal`. JSON output uses snake_case keys and decimal strings.

## Quick start

```go
qty := decimalx.MustFromString("1.5")
price := decimalx.MustFromString("42000")
order, err := domain.NewOrder("BTC-USDT", domain.OrderSideBuy, domain.OrderTypeLimit, qty, price)
if err != nil {
    // handle validation error
}
submitted, err := order.TransitionTo(domain.OrderStateSubmitted)
_ = submitted
```

## Verification

Expected release checks:

```sh
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
```
