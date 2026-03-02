package connection

import (
	"errors"
	"sync"
	"time"
)

var ErrConnectionBroken = errors.New("connection is currently broken")

type Status int

const (
	StatusHealthy Status = iota
	StatusBroken
	StatusTesting
)

type State struct {
	mu              sync.Mutex
	status          Status
	failuresCount   int
	lastFailureTime time.Time
	maxFailures     int
	resetTimeout    time.Duration
}

func New(maxFailures int, resetTimeout time.Duration) *State {
	return &State{maxFailures: maxFailures, resetTimeout: resetTimeout}
}

func (b *State) IsAllowed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.status {
	case StatusHealthy:
		return true
	case StatusBroken:
		if time.Since(b.lastFailureTime) >= b.resetTimeout {
			b.status = StatusTesting
			return true
		}
		return false
	case StatusTesting:
		return true
	}
	return false
}

func (b *State) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = StatusHealthy
	b.failuresCount = 0
}

func (b *State) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failuresCount++
	b.lastFailureTime = time.Now()
	if b.failuresCount >= b.maxFailures {
		b.status = StatusBroken
	}
}
