package client

import (
	"net/http"

	"github.com/f2prateek/train"
	"github.com/pkg/errors"
	"github.com/sony/gobreaker"
)

// NewCircuitBreaker returns an Interceptor for http.Client
// which can be used as RoundTripper middleware
func NewCircuitBreaker(settings gobreaker.Settings) train.Interceptor {
	cb := gobreaker.NewTwoStepCircuitBreaker(settings)
	return &circuitBreakerInterceptor{cb}
}

type circuitBreakerInterceptor struct {
	*gobreaker.TwoStepCircuitBreaker
}

// Intercept Http client call with circuit breaker check
func (cb *circuitBreakerInterceptor) Intercept(chain train.Chain) (*http.Response, error) {
	done, err := cb.Allow()
	if err != nil {
		return nil, errors.Wrap(err, "CircuitBreaker OPEN")
	}

	resp, err := chain.Proceed(chain.Request())
	if err != nil {
		return resp, errors.Errorf("HTTP request failed with error: %s", err)
	}

	done(checkResponse(resp) == nil)

	return resp, err
}

func checkResponse(resp *http.Response) error {
	if c := resp.StatusCode; 200 <= c && c <= 299 || c == http.StatusNotFound {
		return nil
	}

	return errors.Errorf("HTTP Response with status code %d", resp.StatusCode)
}
