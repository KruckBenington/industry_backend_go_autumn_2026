package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {

	var lastTime time.Time
	if clock == nil {
		lastTime = time.Now()
	}

	return &Limiter{
		mu:     sync.Mutex{},
		clock:  clock,
		rate:   ratePerSec,
		burst:  burst,
		tokens: float64(burst),
		last:   lastTime,
	}
}
func (l *Limiter) AllowN(n int) bool {
	if l.burst <= 0 || l.clock == nil || n <= 0 || n > l.burst {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	currentTime := l.clock.Now()
	if !currentTime.Before(l.last) && l.rate > 0 {
		l.tokens = min(l.tokens+(currentTime.Sub(l.last).Seconds())*l.rate, float64(l.burst))
		l.last = currentTime
	}

	if float64(n) > l.tokens {
		return false
	}

	l.tokens -= float64(n)
	return true
}
