package reflect

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// setting is a request option, read from its header or else from its query param.
type setting struct{ header, query string }

// The shaping params are prefixed, since a `status` or `delay` param belongs to the request.
var (
	formatOption   = setting{"X-Glass-Format", "format"}
	statusOption   = setting{"X-Glass-Status", "glass-status"}
	delayOption    = setting{"X-Glass-Delay", "glass-delay"}
	encodingOption = setting{"X-Glass-Encoding", "glass-encoding"}
)

func (s setting) from(r *http.Request) string {
	if v := r.Header.Get(s.header); v != "" {
		return v
	}

	return r.URL.Query().Get(s.query)
}

func (s setting) invalid(want, got string) error {
	return fmt.Errorf("%s (or ?%s=) must be %s, got %q", s.header, s.query, want, got)
}

// answer is how a request asks to be answered.
type answer struct {
	status   int
	delay    time.Duration
	encoding string
}

// shape reads the status, delay and encoding a request asks to be answered with.
func shape(r *http.Request, maxDelay time.Duration) (answer, error) {
	a := answer{status: http.StatusOK}

	if v := statusOption.from(r); v != "" {
		code, err := strconv.Atoi(v)
		if err != nil || code < 200 || code > 599 {
			return a, statusOption.invalid("a status from 200 to 599", v)
		}
		a.status = code
	}

	if v := delayOption.from(r); v != "" {
		d, err := parseDelay(v)
		if err != nil || d < 0 || d > maxDelay {
			return a, delayOption.invalid("a duration from 0 to "+maxDelay.String(), v)
		}
		a.delay = d
	}

	if v := encodingOption.from(r); v != "" {
		if _, ok := codecs[v]; !ok {
			return a, encodingOption.invalid("gzip or deflate", v)
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
