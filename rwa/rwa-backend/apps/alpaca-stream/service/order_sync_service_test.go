package service

import (
	"testing"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/handlers"
)

func TestExtractClientOrderID(t *testing.T) {
	tests := []struct {
		name    string
		data    handlers.TradeUpdateMessageData
		want    string
		wantErr bool
	}{
		{
			name: "valid client_order_id",
			data: handlers.TradeUpdateMessageData{
				Order: handlers.AlpacaOrderData{
					ClientOrderID: "order-123",
				},
			},
			want:    "order-123",
			wantErr: false,
		},
		{
			name: "empty client_order_id",
			data: handlers.TradeUpdateMessageData{
				Order: handlers.AlpacaOrderData{
					ClientOrderID: "",
				},
			},
			wantErr: true,
		},
		{
			name:    "zero value data",
			data:    handlers.TradeUpdateMessageData{},
			wantErr: true,
		},
		{
			name: "whitespace-only client_order_id is accepted (not empty)",
			data: handlers.TradeUpdateMessageData{
				Order: handlers.AlpacaOrderData{
					ClientOrderID: "   ",
				},
			},
			want:    "   ",
			wantErr: false,
		},
		{
			name: "numeric client_order_id",
			data: handlers.TradeUpdateMessageData{
				Order: handlers.AlpacaOrderData{
					ClientOrderID: "123456789",
				},
			},
			want:    "123456789",
			wantErr: false,
		},
		{
			name: "uuid-format client_order_id",
			data: handlers.TradeUpdateMessageData{
				Order: handlers.AlpacaOrderData{
					ClientOrderID: "550e8400-e29b-41d4-a716-446655440000",
				},
			},
			want:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name: "very long client_order_id",
			data: handlers.TradeUpdateMessageData{
				Order: handlers.AlpacaOrderData{
					ClientOrderID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			want:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: false,
		},
		{
			name: "client_order_id with special characters",
			data: handlers.TradeUpdateMessageData{
				Order: handlers.AlpacaOrderData{
					ClientOrderID: "order-123_test.v2",
				},
			},
			want:    "order-123_test.v2",
			wantErr: false,
		},
		{
			name: "other fields populated but client_order_id empty",
			data: handlers.TradeUpdateMessageData{
				Event: "fill",
				Price: "150.00",
				Qty:   "10",
				Order: handlers.AlpacaOrderData{
					ID:            "alpaca-id-123",
					ClientOrderID: "",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractClientOrderID(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("extractClientOrderID() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("extractClientOrderID() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("extractClientOrderID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimestampOrNow(t *testing.T) {
	tests := []struct {
		name       string
		ts         string
		wantExact  *time.Time // if non-nil, check exact match
		wantRecent bool       // if true, expect result close to time.Now()
	}{
		{
			name: "valid RFC3339Nano timestamp",
			ts:   "2024-01-15T10:30:00.123456789Z",
			wantExact: func() *time.Time {
				t, _ := time.Parse(time.RFC3339Nano, "2024-01-15T10:30:00.123456789Z")
				return &t
			}(),
		},
		{
			name: "valid RFC3339 timestamp (no nanos)",
			ts:   "2024-06-01T12:00:00Z",
			wantExact: func() *time.Time {
				t, _ := time.Parse(time.RFC3339Nano, "2024-06-01T12:00:00Z")
				return &t
			}(),
		},
		{
			name:       "empty string falls back to now",
			ts:         "",
			wantRecent: true,
		},
		{
			name:       "invalid format falls back to now",
			ts:         "not-a-timestamp",
			wantRecent: true,
		},
		{
			name:       "partial date falls back to now",
			ts:         "2024-01-15",
			wantRecent: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Add(-1 * time.Second)
			got := parseTimestampOrNow(tt.ts)
			after := time.Now().Add(1 * time.Second)

			if got == nil {
				t.Fatalf("parseTimestampOrNow(%q) returned nil", tt.ts)
			}
			if tt.wantExact != nil {
				if !got.Equal(*tt.wantExact) {
					t.Errorf("parseTimestampOrNow(%q) = %v, want %v", tt.ts, *got, *tt.wantExact)
				}
			}
			if tt.wantRecent {
				if got.Before(before) || got.After(after) {
					t.Errorf("parseTimestampOrNow(%q) = %v, expected close to now (between %v and %v)", tt.ts, *got, before, after)
				}
			}
		})
	}
}
