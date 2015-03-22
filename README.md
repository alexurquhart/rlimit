#rlimit
Package rlimit contains utilities to help with complex rate limiting scenarios.
The RateLimiter struct works by using a combination of a ticker that ticks at a fixed
rate, and a specified limit on the number of ticks allowed before waiting for a reset.
This was developed because an API that my application was consuming allowed querying
at a rate of 50 calls/min, up to a maximum of 1000 calls per hour, and I needed a way to
create long-running processes that could get the data as soon as possible without exceeding
the API's limits.

##Examples
Use the Wait() method to block until the next tick is reached.
```go
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

	now := time.Now()

	// Make a bunch of API calls. The Wait() method will block until
	// the appropriate time has passed.
	for i := 0; i < 15; i++ {
		now = time.Now()
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
```
Outputs
```
Expensive API call blocked for: 250.384505ms
Expensive API call blocked for: 249.747842ms
Expensive API call blocked for: 249.995701ms
Expensive API call blocked for: 249.9797ms
Expensive API call blocked for: 249.982543ms
Expensive API call blocked for: 1.75011375s
Expensive API call blocked for: 249.985946ms
Expensive API call blocked for: 249.982977ms
Expensive API call blocked for: 249.996601ms
Expensive API call blocked for: 249.966797ms
Expensive API call blocked for: 2.000079886s	
Expensive API call blocked for: 249.986876ms
Expensive API call blocked for: 249.98135ms
Expensive API call blocked for: 249.980088ms
Expensive API call blocked for: 249.978802ms
```
Note how once 5 API calls are made the application blocks until the reset interval has passed.

The examples folder contains more demonstrations, including how to use the rate limiter across
multiple goroutines
