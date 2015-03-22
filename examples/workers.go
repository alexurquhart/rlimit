package main

import (
	"fmt"
	"github.com/alexurquhart/rlimit"
	"time"
)

func worker(limiter *rlimit.RateLimiter, id string) {
	for {
		_, err := limiter.Wait()

		// Wait() will return an error when the limiter has been stopped
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("Worker %s:\tLimit Left: %d\n", id, limiter.LimitLeft())
	}
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
