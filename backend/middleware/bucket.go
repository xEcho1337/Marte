package middleware

import (
	"sync"
	"time"
)

type Bucket struct {
	mu         sync.Mutex
	Tokens     float64
	LastRefill time.Time
	Capacity   float64
	RefillRate float64
}

func (b *Bucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.LastRefill).Seconds()
	b.Tokens = min(b.Capacity, b.Tokens+elapsed*b.RefillRate)
	b.LastRefill = now

	if b.Tokens >= 1 {
		b.Tokens--
		return true
	}
	return false
}
