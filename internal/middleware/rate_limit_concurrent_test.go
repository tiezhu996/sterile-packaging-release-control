package middleware

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiterConcurrentScopes(t *testing.T) {
	login := NewNamedRateLimiter(nil, "login", 10, time.Minute)
	global := NewNamedRateLimiter(nil, "global", 180, time.Minute)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			<-start
			login.localCount("rate:login:client-" + string(rune('a'+n%26)))
		}(i)
		go func(n int) {
			defer wg.Done()
			<-start
			global.localCount("rate:global:client-" + string(rune('a'+n%26)))
		}(i)
	}
	close(start)
	wg.Wait()

	if got, _ := login.localCount("rate:login:alone"); got != 1 {
		t.Fatalf("login scope count polluted: got %d, want 1", got)
	}
	if got, _ := global.localCount("rate:global:alone"); got != 1 {
		t.Fatalf("global scope count polluted: got %d, want 1", got)
	}
}
