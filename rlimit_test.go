package rlimit

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	// Create a limiter that ticks every 100 ms and
	// has a limit of 5 ticks before waiting for the reset
	// which happens 1 second after the limiter is created
	r := NewRateLimiter(time.Duration(100)*time.Millisecond, 5, time.Second)

	assert.EqualValues(t, r.LimitLeft(), 5)

	// Simulate a ticker
	_, err := r.Wait()
	assert.NoError(t, err)
	assert.EqualValues(t, r.LimitLeft(), 4)

	// Start blocking in a separate goroutine
	// Expect an error as the ticker is stopped right afterwards
	go func() {
		_, err = r.Wait()
		assert.Error(t, err)
	}()
	r.Stop()
}
