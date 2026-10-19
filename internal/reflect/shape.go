package reflect

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// The query params are prefixed, since a `status` or `delay` param belongs to the request.
const (
	StatusHeader = "X-Glass-Status"
	StatusQuery  = "glass-status"
	DelayHeader  = "X-Glass-Delay"
	DelayQuery   = "glass-delay"
)

// shape reads the status and delay a request asks to be answered with.
func shape(r *http.Request, maxDelay time.Duration) (int, time.Duration, error) {
	status := http.StatusOK
	if v := option(r, StatusHeader, StatusQuery); v != "" {
		code, err := strconv.Atoi(v)
		if err != nil || code < 200 || code > 599 {
			return 0, 0, fmt.Errorf("%s (or ?%s=) must be a status from 200 to 599, got %q", StatusHeader, StatusQuery, v)
		}
		status = code
	}

	var delay time.Duration
	if v := option(r, DelayHeader, DelayQuery); v != "" {
		d, err := parseDelay(v)
		if err != nil || d < 0 || d > maxDelay {
			return 0, 0, fmt.Errorf("%s (or ?%s=) must be a duration from 0 to %s, got %q", DelayHeader, DelayQuery, maxDelay, v)
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
