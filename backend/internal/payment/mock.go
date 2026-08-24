package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gomusical/internal/clock"
	"gomusical/internal/logx"
)

// MockProvider simulates four behaviors: success / fail / timeout / delayed.
type MockProvider struct {
	Behavior string
	Public   string
}

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) Charge(_ context.Context, req ChargeRequest) (ChargeResult, error) {
	beh := strings.ToLower(m.Behavior)
	if beh == "" {
		beh = "success"
	}
	logx.Info("mock charge", "order", req.OrderNo, "behavior", beh, "cents", req.AmountCents)
	switch beh {
	case "fail":
		return ChargeResult{Status: "failed", ProviderTx: "mock_fail_" + req.OrderNo}, fmt.Errorf("mock payment failed")
	case "timeout":
		return ChargeResult{Status: "pending", ProviderTx: "mock_timeout_" + req.OrderNo, CheckoutURL: m.checkout(req.OrderNo, "timeout")}, nil
	case "delayed":
		return ChargeResult{Status: "pending", ProviderTx: "mock_delay_" + req.OrderNo, CheckoutURL: m.checkout(req.OrderNo, "delayed")}, nil
	default:
		now := clock.Now()
		return ChargeResult{
			Status:      "paid",
			ProviderTx:  "mock_ok_" + req.OrderNo,
			CheckoutURL: m.checkout(req.OrderNo, "success"),
			PaidAt:      &now,
		}, nil
	}
}

func (m *MockProvider) VerifyCallback(_ context.Context, payload map[string]string) (string, bool, error) {
	order := payload["orderNo"]
	if order == "" {
		return "", false, fmt.Errorf("missing orderNo")
	}
	if payload["status"] == "failed" {
		return order, false, nil
	}
	return order, true, nil
}

func (m *MockProvider) checkout(orderNo, beh string) string {
	base := strings.TrimRight(m.Public, "/")
	return fmt.Sprintf("%s/pay/mock?orderNo=%s&behavior=%s", base, orderNo, beh)
}

func DelayThenPaid() time.Time {
	return clock.Now().Add(2 * time.Second)
}
