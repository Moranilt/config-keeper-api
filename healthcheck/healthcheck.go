package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Moranilt/http-utils/response"
)

type HealthItem struct {
	Name    string
	Checker Checker
}

type health struct {
	checkers []HealthItem
	timeout  time.Duration
}

type Checker interface {
	Check(context.Context) error
}

func Handler(items ...HealthItem) http.Handler {
	return &health{
		checkers: items,
		timeout:  30 * time.Second,
	}
}

func HandlerFunc(items ...HealthItem) http.HandlerFunc {
	return Handler(items...).ServeHTTP
}

func (h *health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errors := make(map[string]string, len(h.checkers))
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(h.checkers))
	code := http.StatusOK

	for _, checker := range h.checkers {
		go func(item HealthItem) {
			if err := item.Checker.Check(ctx); err != nil {
				mu.Lock()
				errors[item.Name] = err.Error()
				code = http.StatusServiceUnavailable
				mu.Unlock()
			}
			wg.Done()
		}(checker)
	}

	wg.Wait()
	w.Header().Set("Content-Type", "application/json")

	response.Default(w, http.StatusText(code), errors, code)
}
