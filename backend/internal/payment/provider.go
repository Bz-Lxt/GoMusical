package payment

import (
	"context"
	"time"
)

type ChargeRequest struct {
	OrderNo     string
	AmountCents int
	UserID      string
	Subject     string
	ReturnURL   string
}

type ChargeResult struct {
	ProviderTx string
	CheckoutURL string
	Status     string
	PaidAt     *time.Time
}

type Provider interface {
	Name() string
	Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error)
	VerifyCallback(ctx context.Context, payload map[string]string) (orderNo string, ok bool, err error)
}
