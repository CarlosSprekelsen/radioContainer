//go:build integration

package fixtures

import (
	"sync"
	"time"
)

// ManualClock provides deterministic time control for integration tests.
type ManualClock struct {
	mu   sync.RWMutex
	now  time.Time
}

// NewManualClock creates a new manual clock starting at the current time.
func NewManualClock() *ManualClock {
	return &ManualClock{
		now: time.Now(),
	}
}

// NewManualClockAt creates a new manual clock starting at the specified time.
func NewManualClockAt(t time.Time) *ManualClock {
	return &ManualClock{
		now: t,
	}
}

// Now returns the current time according to the manual clock.
func (c *ManualClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

// Advance moves the clock forward by the specified duration.
func (c *ManualClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// Set sets the clock to the specified time.
func (c *ManualClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

// Since returns the duration since the specified time according to the manual clock.
func (c *ManualClock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}
