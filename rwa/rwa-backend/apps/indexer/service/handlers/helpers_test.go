package handlers

import (
	"math/big"
	"testing"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

func TestConvertUint8ToOrderSide(t *testing.T) {
	tests := []struct {
		name string
		side uint8
		want rwa.OrderSide
	}{
		{"0 -> buy", 0, rwa.OrderSideBuy},
		{"1 -> sell", 1, rwa.OrderSideSell},
		{"2 -> sell (default)", 2, rwa.OrderSideSell},
		{"255 -> sell (default)", 255, rwa.OrderSideSell},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertUint8ToOrderSide(tt.side)
			if got != tt.want {
				t.Errorf("convertUint8ToOrderSide(%d) = %v, want %v", tt.side, got, tt.want)
			}
		})
	}
}

func TestConvertUint8ToOrderType(t *testing.T) {
	tests := []struct {
		name      string
		orderType uint8
		want      rwa.OrderType
	}{
		{"0 -> market", 0, rwa.OrderTypeMarket},
		{"1 -> limit", 1, rwa.OrderTypeLimit},
		{"2 -> stop", 2, rwa.OrderTypeStop},
		{"3 -> stop_limit", 3, rwa.OrderTypeStopLimit},
		{"4 -> market (default)", 4, rwa.OrderTypeMarket},
		{"255 -> market (default)", 255, rwa.OrderTypeMarket},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertUint8ToOrderType(tt.orderType)
			if got != tt.want {
				t.Errorf("convertUint8ToOrderType(%d) = %v, want %v", tt.orderType, got, tt.want)
			}
		})
	}
}

func TestConvertUint8ToOrderStatus(t *testing.T) {
	tests := []struct {
		name   string
		status uint8
		want   rwa.OrderStatus
	}{
		{"0 -> new", 0, rwa.OrderStatusNew},
		{"1 -> pending", 1, rwa.OrderStatusPending},
		{"2 -> accepted", 2, rwa.OrderStatusAccepted},
		{"3 -> rejected", 3, rwa.OrderStatusRejected},
		{"4 -> filled", 4, rwa.OrderStatusFilled},
		{"5 -> partially_filled", 5, rwa.OrderStatusPartiallyFilled},
		{"6 -> cancelled", 6, rwa.OrderStatusCancelled},
		{"7 -> expired", 7, rwa.OrderStatusExpired},
		{"8 -> new (default)", 8, rwa.OrderStatusNew},
		{"255 -> new (default)", 255, rwa.OrderStatusNew},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertUint8ToOrderStatus(tt.status)
			if got != tt.want {
				t.Errorf("convertUint8ToOrderStatus(%d) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestConvertUint8ToTimeInForce(t *testing.T) {
	tests := []struct {
		name string
		tif  uint8
		want alpaca.TimeInForce
	}{
		{"0 -> Day", 0, alpaca.Day},
		{"1 -> GTC", 1, alpaca.GTC},
		{"2 -> IOC", 2, alpaca.IOC},
		{"3 -> FOK", 3, alpaca.FOK},
		{"4 -> OPG", 4, alpaca.OPG},
		{"5 -> Day (GTX fallback)", 5, alpaca.Day},
		{"6 -> GTC (GTD fallback)", 6, alpaca.GTC},
		{"7 -> CLS", 7, alpaca.CLS},
		{"8 -> Day (default)", 8, alpaca.Day},
		{"255 -> Day (default)", 255, alpaca.Day},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertUint8ToTimeInForce(tt.tif)
			if got != tt.want {
				t.Errorf("convertUint8ToTimeInForce(%d) = %v, want %v", tt.tif, got, tt.want)
			}
		})
	}
}

func TestBigIntToDecimalWithPrecision(t *testing.T) {
	tests := []struct {
		name     string
		bi       *big.Int
		decimals int32
		want     string
	}{
		{"nil returns zero", nil, 18, "0"},
		{"zero value", big.NewInt(0), 18, "0"},
		{"1.5 with 18 decimals", new(big.Int).SetUint64(1500000000000000000), 18, "1.5"},
		{"1 with 18 decimals", new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), 18, "1"},
		{"1 USDC (6 decimals)", big.NewInt(1000000), 6, "1"},
		{"0.5 USDC (6 decimals)", big.NewInt(500000), 6, "0.5"},
		{"negative value", big.NewInt(-1500000000000000000), 18, "-1.5"},
		{"zero decimals", big.NewInt(42), 0, "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bigIntToDecimalWithPrecision(tt.bi, tt.decimals)
			want, _ := decimal.NewFromString(tt.want)
			if !got.Equal(want) {
				t.Errorf("bigIntToDecimalWithPrecision(%v, %d) = %v, want %v", tt.bi, tt.decimals, got, want)
			}
		})
	}
}

func TestParseBlockTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "valid hex timestamp",
			input:   "0x60000000",
			want:    time.Unix(0x60000000, 0),
			wantErr: false,
		},
		{
			name:    "zero timestamp",
			input:   "0x0",
			want:    time.Unix(0, 0),
			wantErr: false,
		},
		{
			name:    "invalid hex string",
			input:   "not-a-hex",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBlockTimestamp(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseBlockTimestamp(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseBlockTimestamp(%q) unexpected error: %v", tt.input, err)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("parseBlockTimestamp(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
