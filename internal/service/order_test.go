package service

import (
	"testing"

	"hotel-backend/internal/model"
)

func TestCancelTargetForStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		wantStatus    string
		wantImmediate bool
		wantErr       error
	}{
		{
			name:          "pending orders cancel immediately",
			status:        model.OrderStatusPending,
			wantStatus:    model.OrderStatusCancelled,
			wantImmediate: true,
		},
		{
			name:          "confirmed orders enter cancel review",
			status:        model.OrderStatusConfirmed,
			wantStatus:    model.OrderStatusCancelRequested,
			wantImmediate: false,
		},
		{
			name:    "cancel requested orders cannot be cancelled again",
			status:  model.OrderStatusCancelRequested,
			wantErr: ErrOrderNotPending,
		},
		{
			name:    "cancelled orders cannot be cancelled again",
			status:  model.OrderStatusCancelled,
			wantErr: ErrOrderNotPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotImmediate, gotErr := cancelTargetForStatus(tt.status)
			if gotErr != tt.wantErr {
				t.Fatalf("error = %v, want %v", gotErr, tt.wantErr)
			}
			if gotStatus != tt.wantStatus {
				t.Fatalf("status = %q, want %q", gotStatus, tt.wantStatus)
			}
			if gotImmediate != tt.wantImmediate {
				t.Fatalf("immediate = %v, want %v", gotImmediate, tt.wantImmediate)
			}
		})
	}
}

func TestValidateOrderUpdateStatus(t *testing.T) {
	tests := []struct {
		name      string
		oldStatus string
		newStatus string
		wantErr   bool
	}{
		{
			name:      "cannot manually create cancel review",
			oldStatus: model.OrderStatusPending,
			newStatus: model.OrderStatusCancelRequested,
			wantErr:   true,
		},
		{
			name:      "cannot directly cancel normal pending order",
			oldStatus: model.OrderStatusPending,
			newStatus: model.OrderStatusCancelled,
			wantErr:   true,
		},
		{
			name:      "cannot directly cancel normal confirmed order",
			oldStatus: model.OrderStatusConfirmed,
			newStatus: model.OrderStatusCancelled,
			wantErr:   true,
		},
		{
			name:      "cancel review can be approved",
			oldStatus: model.OrderStatusCancelRequested,
			newStatus: model.OrderStatusCancelled,
		},
		{
			name:      "cancel review can be rejected back to confirmed",
			oldStatus: model.OrderStatusCancelRequested,
			newStatus: model.OrderStatusConfirmed,
		},
		{
			name:      "pending order can be confirmed",
			oldStatus: model.OrderStatusPending,
			newStatus: model.OrderStatusConfirmed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOrderUpdateStatus(tt.oldStatus, tt.newStatus)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
