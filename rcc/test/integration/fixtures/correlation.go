//go:build integration

package fixtures

import (
	"fmt"
	"sync/atomic"
)

// CorrelationIDGenerator provides deterministic correlation IDs for integration tests.
type CorrelationIDGenerator struct {
	counter int64
}

// NewCorrelationIDGenerator creates a new deterministic correlation ID generator.
func NewCorrelationIDGenerator() *CorrelationIDGenerator {
	return &CorrelationIDGenerator{}
}

// Next returns the next deterministic correlation ID.
// Returns "fixed-1", "fixed-2", "fixed-3", etc.
func (g *CorrelationIDGenerator) Next() string {
	n := atomic.AddInt64(&g.counter, 1)
	return fmt.Sprintf("fixed-%d", n)
}

// Reset resets the counter to start from "fixed-1" again.
func (g *CorrelationIDGenerator) Reset() {
	atomic.StoreInt64(&g.counter, 0)
}
