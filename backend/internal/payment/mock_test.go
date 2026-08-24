package payment

import (
	"context"
	"testing"
)

func TestMockSuccess(t *testing.T) {
	m := &MockProvider{Behavior: "success", Public: "http://localhost:29471"}
	res, err := m.Charge(context.Background(), ChargeRequest{OrderNo: "GM1", AmountCents: 900})
	if err != nil || res.Status != "paid" || res.PaidAt == nil {
		t.Fatalf("%v %+v", err, res)
	}
}

func TestMockFail(t *testing.T) {
	m := &MockProvider{Behavior: "fail"}
	_, err := m.Charge(context.Background(), ChargeRequest{OrderNo: "GM2", AmountCents: 1})
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestCallback(t *testing.T) {
	m := &MockProvider{Behavior: "success"}
	no, ok, err := m.VerifyCallback(context.Background(), map[string]string{"orderNo": "GM3", "status": "paid"})
	if err != nil || !ok || no != "GM3" {
		t.Fatal("callback")
	}
	_, ok, _ = m.VerifyCallback(context.Background(), map[string]string{"orderNo": "GM3", "status": "failed"})
	if ok {
		t.Fatal("failed should not verify")
	}
}

func TestMemoryGate(t *testing.T) {
	g := NewMemoryGate()
	first := g.Remember("A", ChargeResult{Status: "paid"})
	second := g.Remember("A", ChargeResult{Status: "failed"})
	if first.Status != second.Status || second.Status != "paid" {
		t.Fatal("gate must return first result")
	}
}

func TestSelectDegrade(t *testing.T) {
	p := Select("real", "", "http://x", "success")
	if p.Name() != "mock" {
		t.Fatal("must degrade")
	}
}
