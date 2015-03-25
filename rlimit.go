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

type limits struct {
	sync.RWMutex

	// Maximum number of tick that will be allowed until
	// the limiter needs to wait until a reset
	limit uint

	// Variable that holds the # of ticks left until it hits the limit
	limitLeft uint

	// The duration needed to pass before the limit is reset
	resetInterval time.Duration
}

func (l *limits) SetLimitLeft(i uint) {
	l.Lock()
	l.limitLeft = i
	l.Unlock()
}

func (l *limits) DecrLimitLeft() {
	l.Lock()
	l.limitLeft -= 1
	l.Unlock()
}

func (l *limits) LimitLeft() uint {
	l.RLock()
	defer l.RUnlock()
	return l.limitLeft
}

func (l *limits) ResetInterval() time.Duration {
	l.RLock()
	defer l.RUnlock()
	return l.resetInterval
}

type RateLimiter struct {
	// Ticker that will be set to tick at the given interval
	ticker *time.Ticker

	// Timer that will signal when the limits will be reset
	resetTimer *time.Timer

	// Channel to carry the stop event
	stop chan bool

	// Rate limit infomation - protected by a RWMutex
	limits *limits

	// Channel to carry tick events
	Tick chan time.Time
}

// Creates a new rate limiter.
func NewRateLimiter(tickInterval time.Duration, limit uint, resetInterval time.Duration) *RateLimiter {
	l := &limits{
		limit:         limit,
		limitLeft:     limit,
		resetInterval: resetInterval,
	}

	r := &RateLimiter{
		ticker:     time.NewTicker(tickInterval),
		resetTimer: time.NewTimer(resetInterval),
		stop:       make(chan bool),
		limits:     l,
		Tick:       make(chan time.Time),
	}

	// Start a goroutine that manages the state
	// of the tickers, timers, and limits.
	go func() {
		for {
			select {
			case t := <-r.ticker.C:
				if r.limits.LimitLeft() > 0 {
					r.Tick <- t
				} else {
					// Wait for the reset timer
					<-r.resetTimer.C

					// Reset the timer and the limit
					r.resetTimer.Reset(l.resetInterval)
					r.limits.SetLimitLeft(r.limits.limit)
				}
			case <-r.stop:
				close(r.Tick)
				return
			}
		}
	}()

	return r
}

// Stops the internal ticker and closes the Tick channel
func (r *RateLimiter) Stop() {
	r.ticker.Stop()
	close(r.stop)
}

// Blocks until the next tick. Returns the time of the tick, and an error
// if the rate limiter has been stopped/the tick channel has closed
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
	r.limits.DecrLimitLeft()
}

// Returns the number of ticks left until the limiter blocks and
// waits for the reset timer.
func (r *RateLimiter) LimitLeft() uint {
	return r.limits.LimitLeft()
}
