package payment

import "sync"

// MemoryGate drops duplicate Charge calls for the same orderNo in-process.
// Durable idempotency still lives in orders.status via Fulfill.
type MemoryGate struct {
	mu   sync.Mutex
	seen map[string]ChargeResult
}

func NewMemoryGate() *MemoryGate {
	return &MemoryGate{seen: map[string]ChargeResult{}}
}

func (g *MemoryGate) Remember(orderNo string, res ChargeResult) ChargeResult {
	g.mu.Lock()
	defer g.mu.Unlock()
	if prev, ok := g.seen[orderNo]; ok {
		return prev
	}
	g.seen[orderNo] = res
	return res
}

func (g *MemoryGate) Has(orderNo string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, ok := g.seen[orderNo]
	return ok
}
