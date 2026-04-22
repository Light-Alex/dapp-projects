package trade

import (
	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/shopspring/decimal"
)

// validateOrderRequest validates order request parameters
func validateOrderRequest(symbol, side, orderType string, quantity decimal.Decimal) error {
	if symbol == "" {
		return errors.Errorf("symbol is required")
	}
	if side == "" {
		return errors.Errorf("side is required")
	}
	if orderType == "" {
		return errors.Errorf("order type is required")
	}
	if quantity.LessThanOrEqual(decimal.Zero) {
		return errors.Errorf("quantity must be greater than zero")
	}
	return nil
}

// validateSymbol validates symbol format
func validateSymbol(symbol string) error {
	if symbol == "" {
		return errors.Errorf("symbol is required")
	}
	// Add more validation rules if needed
	return nil
}

// normalizeInterval normalizes interval format to standard format
// Supports two formats only:
//   - Short format: 1m, 5m, 15m, 1h, 1d, 1w
//   - Full format: 1Min, 5Min, 15Min, 1Hour, 1Day, 1Week
func normalizeInterval(interval string) (string, error) {
	switch interval {
	// Minute intervals
	case "1m", "1Min":
		return "1m", nil
	case "5m", "5Min":
		return "5m", nil
	case "15m", "15Min":
		return "15m", nil
	// Hour intervals
	case "1h", "1Hour":
		return "1h", nil
	// Day intervals
	case "1d", "1Day":
		return "1d", nil
	// Week intervals
	case "1w", "1Week":
		return "1w", nil
	default:
		return "", errors.Errorf("unsupported time interval: %s", interval)
	}
}
