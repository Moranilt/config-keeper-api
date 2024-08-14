package callback

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Moranilt/http-utils/client"
	"github.com/Moranilt/http-utils/logger"
)

const (
	MAX_RETRIES     = 3
	BASE_DELAY      = 100 * time.Millisecond
	MAX_DELAY       = 5 * time.Second
	REQUEST_TIMEOUT = 10 * time.Second
)

type RequestsController interface {
	SendRequestWithRetry(ctx context.Context, endpoint string, data []byte) error
}

type requestsController struct {
	log        logger.Logger
	httpClient client.Client
}

func NewRequestsController(
	log logger.Logger,
	httpClient client.Client,
) RequestsController {
	return &requestsController{
		log:        log,
		httpClient: httpClient,
	}
}

func (s *requestsController) SendRequestWithRetry(ctx context.Context, endpoint string, data []byte) error {
	for attempt := 0; attempt < MAX_RETRIES; attempt++ {
		err := s.sendRequest(ctx, endpoint, data)
		if err == nil {
			return nil
		}

		delay := s.calculateBackoff(attempt, BASE_DELAY, MAX_DELAY)

		select {
		case <-time.After(delay):
			s.log.Debugf("retrying request to %s in %s", endpoint, delay)
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("max retries reached for endpoint %s", endpoint)
}

func (s *requestsController) sendRequest(ctx context.Context, endpoint string, data []byte) error {
	ctx, cancel := context.WithTimeout(ctx, REQUEST_TIMEOUT)
	defer cancel()

	resp, err := s.httpClient.Post(ctx, endpoint, data, client.NewHeaders(map[string]string{
		"Content-Type": "application/json",
	}))
	if err != nil {
		s.log.Errorf("error sending callback to %s: %s", endpoint, err)
		return err
	}
	if resp.StatusCode >= 500 {
		s.log.Errorf("server error from %s: %d", endpoint, resp.StatusCode)
		return fmt.Errorf("server error from %s: %d", endpoint, resp.StatusCode)
	}
	s.log.Infof("Callback sent to %s with status code %d", endpoint, resp.StatusCode)
	return nil
}

func (s *requestsController) calculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}
