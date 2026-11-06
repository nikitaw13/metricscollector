package retry

import (
	"net/http"
	"slices"
	"time"

	"go.uber.org/zap"
)

// HTTPClient abstracts a single-request HTTP client; *http.Client implements it.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var _ HTTPClient = (*ClientWithRetries)(nil)

// ClientWithRetries wraps an HTTPClient and retries transport failures with the configured backoff.
type ClientWithRetries struct {
	retryDelays []time.Duration
	client      HTTPClient
	logger      *zap.Logger
}

// New wraps client with the given retryDelays between attempts.
func New(retryDelays []time.Duration, client HTTPClient, logger *zap.Logger) *ClientWithRetries {
	return &ClientWithRetries{
		client:      client,
		retryDelays: slices.Clone(retryDelays),
		logger:      logger,
	}
}

// Do sends the request, retrying transport failures with the configured backoff between attempts.
func (c *ClientWithRetries) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)

	if err != nil {
		if req.GetBody == nil {
			return nil, err
		}

		c.logger.Debug("sending failed", zap.Error(err))

		for i := 0; i < len(c.retryDelays); i++ {
			time.Sleep(c.retryDelays[i])

			req.Body, err = req.GetBody()

			if err != nil {
				return nil, err
			}
			resp, err = c.client.Do(req)
			if err != nil {
				c.logger.Debug("retry failed", zap.Int("attempt", i+1), zap.Int("totalAttempts", len(c.retryDelays)), zap.Error(err))
				continue
			}
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
