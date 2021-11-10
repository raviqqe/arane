package main

import "go.uber.org/ratelimit"

type hostThrottler struct {
	limiter     ratelimit.Limiter
	connections semaphore
}

func newHostThrottler(requestPerSecond, maxConnectionsPerHost int) *hostThrottler {
	l := ratelimit.NewUnlimited()

	if requestPerSecond > 0 {
		l = ratelimit.New(requestPerSecond)
	}

	return &hostThrottler{l, newSemaphore(maxConnectionsPerHost)}
}

func (t *hostThrottler) Request() {
	t.limiter.Take()
	t.connections.Request()
}

func (t *hostThrottler) Release() {
	t.connections.Release()
}
