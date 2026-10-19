package reflect

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	StatusHeader = "X-Glass-Status"
	DelayHeader  = "X-Glass-Delay"
)

// shape reads the status and delay a request asks to be answered with.
// Both are headers only: a query param named `status` or `delay` belongs to the request.
func shape(r *http.Request, maxDelay time.Duration) (int, time.Duration, error) {
	status := http.StatusOK
	if v := r.Header.Get(StatusHeader); v != "" {
		code, err := strconv.Atoi(v)
		if err != nil || code < 200 || code > 599 {
			return 0, 0, fmt.Errorf("%s must be a status from 200 to 599, got %q", StatusHeader, v)
		}
		status = code
	}

	var delay time.Duration
	if v := r.Header.Get(DelayHeader); v != "" {
		d, err := parseDelay(v)
		if err != nil || d < 0 || d > maxDelay {
			return 0, 0, fmt.Errorf("%s must be a duration from 0 to %s, got %q", DelayHeader, maxDelay, v)
		}
		delay = d
	}

	return status, delay, nil
}

// parseDelay takes a Go duration (`1.5s`, `300ms`) or a number of seconds.
func parseDelay(v string) (time.Duration, error) {
	if d, err := time.ParseDuration(v); err == nil {
		return d, nil
	}

	seconds, err := strconv.ParseFloat(v, 64)

	return time.Duration(seconds * float64(time.Second)), err
}
