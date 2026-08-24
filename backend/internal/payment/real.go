package payment

import (
	"context"
	"fmt"

	"gomusical/internal/logx"
)

// RealProvider is a skeleton for Alipay / WeChat / Stripe.
// Without PAYMENT_REAL_KEY the process MUST fall back to Mock (config.Load).
type RealProvider struct {
	Key    string
	Public string
}

func (r *RealProvider) Name() string { return "real" }

func (r *RealProvider) Charge(_ context.Context, req ChargeRequest) (ChargeResult, error) {
	if r.Key == "" {
		logx.Warn("real payment key missing, refuse charge")
		return ChargeResult{}, fmt.Errorf("PAYMENT_REAL_KEY missing")
	}
	// Skeleton: do not invent a third-party response schema.
	return ChargeResult{}, fmt.Errorf("real payment channel not wired (UNVERIFIED contract)")
}

func (r *RealProvider) VerifyCallback(_ context.Context, _ map[string]string) (string, bool, error) {
	return "", false, fmt.Errorf("real payment callback unverified")
}

func Select(mode, key, public, mockBeh string) Provider {
	if mode == "real" && key != "" {
		logx.Info("payment provider", "name", "real")
		return &RealProvider{Key: key, Public: public}
	}
	if mode == "real" && key == "" {
		logx.Warn("PAYMENT_MODE=real but key empty; degrading to mock")
	}
	logx.Info("payment provider", "name", "mock", "behavior", mockBeh)
	return &MockProvider{Behavior: mockBeh, Public: public}
}
