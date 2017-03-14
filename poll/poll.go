package poll

import (
	"context"
	"errors"
	"net/http"
	"time"
	"fmt"
	"net/url"
)

type UnhealthyError struct {
	url     string
	attempts int
}

func (u UnhealthyError) Error() string {
	return fmt.Sprintf("healthcheck unhealthy (url '%s' was not healthy after %d attempts)", u.url, u.attempts)
}

func Poll(ctx context.Context, healthcheckURL string, pollingFrequency time.Duration) error {
	if _, err := url.Parse(healthcheckURL); err != nil {
		return err
	}

	ticker := time.NewTicker(pollingFrequency)
	defer ticker.Stop()

	attempts := 0

	for {
		select {
		case <-ticker.C:
			if urlHealthy(ctx, healthcheckURL) {
				return nil
			}
			attempts++
		case <-ctx.Done():
			return UnhealthyError{url: healthcheckURL, attempts: attempts}
		}
	}
	return errors.New("Polling failed unexpectedly")
}

func urlHealthy(ctx context.Context, url string) bool {
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
	}

	request, _ := http.NewRequest("GET", url, nil)
	ctxRequest := request.WithContext(ctx)

	response, err := client.Do(ctxRequest)
	if err != nil {
		return false
	}

	return response.StatusCode == http.StatusOK
}
