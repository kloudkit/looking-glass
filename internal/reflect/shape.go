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

// answer is how a request asks to be answered.
type answer struct {
	status   int
	delay    time.Duration
	encoding string
}

// shape reads the status, delay and encoding a request asks to be answered with.
func shape(r *http.Request, maxDelay time.Duration) (answer, error) {
	a := answer{status: http.StatusOK}

	if v := option(r, StatusHeader, StatusQuery); v != "" {
		code, err := strconv.Atoi(v)
		if err != nil || code < 200 || code > 599 {
			return a, fmt.Errorf("%s (or ?%s=) must be a status from 200 to 599, got %q", StatusHeader, StatusQuery, v)
		}
		a.status = code
	}

	if v := option(r, DelayHeader, DelayQuery); v != "" {
		d, err := parseDelay(v)
		if err != nil || d < 0 || d > maxDelay {
			return a, fmt.Errorf("%s (or ?%s=) must be a duration from 0 to %s, got %q", DelayHeader, DelayQuery, maxDelay, v)
		}
		a.delay = d
	}

	if v := option(r, EncodingHeader, EncodingQuery); v != "" {
		if !validEncoding(v) {
			return a, fmt.Errorf("%s (or ?%s=) must be gzip or deflate, got %q", EncodingHeader, EncodingQuery, v)
		}
		a.encoding = v
	}

	return a, nil
}

// parseDelay takes a Go duration (`1.5s`, `300ms`) or a number of seconds.
func parseDelay(v string) (time.Duration, error) {
	if d, err := time.ParseDuration(v); err == nil {
		return d, nil
	}

	seconds, err := strconv.ParseFloat(v, 64)

	return time.Duration(seconds * float64(time.Second)), err
}
