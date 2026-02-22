package ratelimit

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrConcurrentStream  = errors.New("only one active stream per user is allowed")
	ErrUserMinuteLimit   = errors.New("per-user rate limit exceeded: max 10 messages per minute")
	ErrUserHourLimit     = errors.New("per-user rate limit exceeded: max 100 messages per hour")
	ErrTenantMinuteLimit = errors.New("per-tenant rate limit exceeded: max 50 messages per minute")
)

const cleanupInterval = 5 * time.Minute

type Limiter struct {
	mu            sync.Mutex
	activeStreams map[string]bool
	userMinute    map[string]*slidingWindow
	userHour      map[string]*slidingWindow
	tenantMinute  map[string]*slidingWindow
	stopCleanup   chan struct{}
}

type slidingWindow struct {
	timestamps []time.Time
	window     time.Duration
	limit      int
}

func newSlidingWindow(window time.Duration, limit int) *slidingWindow {
	return &slidingWindow{
		window: window,
		limit:  limit,
	}
}

func (sw *slidingWindow) allow(now time.Time) bool {
	cutoff := now.Add(-sw.window)
	clean := sw.timestamps[:0]
	for _, ts := range sw.timestamps {
		if ts.After(cutoff) {
			clean = append(clean, ts)
		}
	}
	sw.timestamps = clean
	return len(sw.timestamps) < sw.limit
}

func (sw *slidingWindow) record(now time.Time) {
	sw.timestamps = append(sw.timestamps, now)
}

func (sw *slidingWindow) isEmpty() bool {
	return len(sw.timestamps) == 0
}

func NewLimiter() *Limiter {
	l := &Limiter{
		activeStreams: make(map[string]bool),
		userMinute:    make(map[string]*slidingWindow),
		userHour:      make(map[string]*slidingWindow),
		tenantMinute:  make(map[string]*slidingWindow),
		stopCleanup:   make(chan struct{}),
	}
	go l.periodicCleanup()
	return l
}

func (l *Limiter) Stop() {
	close(l.stopCleanup)
}

func (l *Limiter) AcquireStream(userID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.activeStreams[userID] {
		return ErrConcurrentStream
	}
	l.activeStreams[userID] = true
	return nil
}

func (l *Limiter) ReleaseStream(userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.activeStreams, userID)
}

func (l *Limiter) AllowMessage(userID, tenantID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	if sw := l.getUserMinute(userID); !sw.allow(now) {
		return ErrUserMinuteLimit
	}
	if sw := l.getUserHour(userID); !sw.allow(now) {
		return ErrUserHourLimit
	}
	if sw := l.getTenantMinute(tenantID); !sw.allow(now) {
		return ErrTenantMinuteLimit
	}

	l.getUserMinute(userID).record(now)
	l.getUserHour(userID).record(now)
	l.getTenantMinute(tenantID).record(now)

	return nil
}

func (l *Limiter) getUserMinute(userID string) *slidingWindow {
	if _, ok := l.userMinute[userID]; !ok {
		l.userMinute[userID] = newSlidingWindow(time.Minute, 10)
	}
	return l.userMinute[userID]
}

func (l *Limiter) getUserHour(userID string) *slidingWindow {
	if _, ok := l.userHour[userID]; !ok {
		l.userHour[userID] = newSlidingWindow(time.Hour, 100)
	}
	return l.userHour[userID]
}

func (l *Limiter) getTenantMinute(tenantID string) *slidingWindow {
	if _, ok := l.tenantMinute[tenantID]; !ok {
		l.tenantMinute[tenantID] = newSlidingWindow(time.Minute, 50)
	}
	return l.tenantMinute[tenantID]
}

func (l *Limiter) periodicCleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-l.stopCleanup:
			return
		case <-ticker.C:
			l.cleanup()
		}
	}
}

func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	evictStaleEntries(l.userMinute, now)
	evictStaleEntries(l.userHour, now)
	evictStaleEntries(l.tenantMinute, now)
}

func evictStaleEntries(m map[string]*slidingWindow, now time.Time) {
	for key, sw := range m {
		sw.allow(now)
		if sw.isEmpty() {
			delete(m, key)
		}
	}
}
