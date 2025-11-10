package expvar

import (
	"expvar"
	"fmt"
	"time"
)

var (
	httpRequests *expvar.Map
)

func init() {
	// this panic if the map already exists
	httpRequests = expvar.NewMap("http_requests")
}

type ExpvarMetrics struct{}

func NewExpvarMetrics() *ExpvarMetrics {
	return &ExpvarMetrics{}
}

func requestToKeyCount(pattern string, code int) string {
	return fmt.Sprintf("%s.%d.count", pattern, code)
}

func requestToKeyDuration(pattern string, code int) string {
	return fmt.Sprintf("%s.%d.duration_ms", pattern, code)
}

func (em *ExpvarMetrics) AddRequestHit(pattern string, code int, duration time.Duration) error {
	httpRequests.Add(requestToKeyCount(pattern, code), 1)
	httpRequests.Add(requestToKeyDuration(pattern, code), duration.Milliseconds())
	return nil

}
