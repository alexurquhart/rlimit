package main

import (
	"fmt"
	"github.com/alexurquhart/rlimit"
	"time"
)

func main() {
	// Create a new limiter that ticks every 250ms, limited to 5 times every 3 seconds
	interval := time.Duration(250) * time.Millisecond
	resetInterval := time.Duration(3) * time.Second
	limiter := rlimit.NewRateLimiter(interval, 5, resetInterval)

	// Make a bunch of limited API calls. The Wait() method will block until
	// the appropriate time has passed.
	for i := 0; i < 15; i++ {
		now := time.Now()
		_, err := limiter.Wait()
		diff := time.Now().Sub(now)

		// Wait() will return an error when the limiter has been stopped
		if err != nil {
			fmt.Println("Rate Limiter Stopped: ", err)
			break
		}
		fmt.Printf("Expensive API call blocked for: %s\n", diff)
	}
}
