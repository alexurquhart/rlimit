// Package rlimit contains utilities to help with complex rate limiting scenarios.
// The RateLimiter struct works by using a combination of a ticker that ticks at a fixed
// rate, and a specified limit on the number of ticks allowed before waiting for a reset.
// This was developed because an API that my application was consuming allowed querying
// at a rate of 50 calls/min, up to a maximum of 1000 calls per hour, and I needed a way to
// create long-running processes that could get the data as soon as possible without exceeding
// the API's limits.
package rlimit

import (
	"errors"
	"sync"
	"time"
)

type RateLimiter struct {
	sync.RWMutex

	// Ticker that will be set to tick at the given interval
	ticker *time.Ticker

	// Timer that will signal when the limits will be reset
	resetTimer *time.Timer

	// Interval that the limiter will tick at until it hits the limit
	tickInterval time.Duration

	// Maximum number of tick that will be allowed until
	// the limiter needs to wait until a reset
	limit uint

	// Variable that holds the # of ticks left until it hits the limit
	limitLeft uint

	// The duration needed to pass before the limit is reset
	resetInterval time.Duration

	// Channel to carry the stop event
	stop chan bool

	// Channel to carry tick events
	Tick chan time.Time
}

// Creates a new rate limiter.
func NewRateLimiter(tickInterval time.Duration, limit uint, resetInterval time.Duration) *RateLimiter {
	l := &RateLimiter{
		ticker:        time.NewTicker(tickInterval),
		resetTimer:    time.NewTimer(resetInterval),
		tickInterval:  tickInterval,
		limit:         limit,
		limitLeft:     limit,
		resetInterval: resetInterval,
		stop:          make(chan bool),
		Tick:          make(chan time.Time),
	}

	// Start a goroutine that manages the state
	// of the tickers, timers, and limits.
	go func() {
		for {
			select {
			case t := <-l.ticker.C:
				if l.limitLeft > 0 {
					l.Tick <- t
				} else {
					l.RLock()

					// Wait for the reset timer
					<-l.resetTimer.C

					// Reset the timer and the limit
					l.resetTimer.Reset(l.resetInterval)
					l.limitLeft = l.limit
					l.RUnlock()
				}
			case <-l.stop:
				close(l.Tick)
				return
			}
		}
	}()

	return l
}

func (r *RateLimiter) Stop() {
	r.ticker.Stop()
	close(r.stop)
}

// Blocks until the next tick
func (r *RateLimiter) Wait() (time.Time, error) {
	if t, ok := <-r.Tick; ok {
		r.Count()
		return t, nil
	} else {
		return time.Time{}, errors.New("Rate limiter has stopped")
	}
}

// Decrements the limit left - not to be used when waiting for a tick using Wait()
func (r *RateLimiter) Count() {
	r.RLock()
	r.limitLeft -= 1
	r.RUnlock()
}

// Returns the number of ticks left until waiting for a reset
func (r *RateLimiter) LimitLeft() uint {
	return r.limitLeft
}
