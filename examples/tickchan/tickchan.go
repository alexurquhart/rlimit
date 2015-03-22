package main

import (
	"fmt"
	"github.com/alexurquhart/rlimit"
	"time"
)

// Utilizing the Tick channel for receiving tick events is
// more idiomatic than using the Wait() method when using the
// RateLimiter in multiple goroutines. Please note you
// need to call Count() in order to decrement the limit counter
// on each iteration.
func worker(limiter *rlimit.RateLimiter, id string) {
	for range limiter.Tick {
		// When using the Tick channel - you need to manually adjust the limit
		limiter.Count()

		fmt.Printf("Worker %s:\tLimit Left: %d\n", id, limiter.LimitLeft())
	}
	fmt.Println("Worker Closing")
}

func main() {
	// Create a new limiter that ticks every 250ms, limited to 5 times every 3 seconds
	interval := time.Duration(250) * time.Millisecond
	resetInterval := time.Duration(3) * time.Second
	limiter := rlimit.NewRateLimiter(interval, 5, resetInterval)

	// Start up 6 workers
	for i := 1; i < 6; i++ {
		go worker(limiter, fmt.Sprint(i))
	}

	// Sleep for a bit
	time.Sleep(time.Duration(10) * time.Second)

	// Stop the workers
	limiter.Stop()

	// Wait for the error messages
	time.Sleep(time.Second)
}
