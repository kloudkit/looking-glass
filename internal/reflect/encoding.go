package reflect

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

// codec is a Content-Encoding the standard library can read and write.
type codec struct {
	reader func(io.Reader) (io.ReadCloser, error)
	writer func(io.Writer) io.WriteCloser
}

var codecs = map[string]codec{
	"gzip": {
		reader: func(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) },
		writer: func(w io.Writer) io.WriteCloser { return gzip.NewWriter(w) },
	},
	"deflate": {
		reader: zlib.NewReader,
		writer: func(w io.Writer) io.WriteCloser { return zlib.NewWriter(w) },
	},
}

// decompress decodes a gzip or deflate body, keeping at most limit bytes so a small upload
// can't expand past the body limit. Any other encoding, or a stream that doesn't decode, is
// left as it arrived.
func decompress(encoding string, data []byte, limit int64) ([]byte, string, bool) {
	encoding = strings.ToLower(strings.TrimSpace(encoding))
	if encoding == "x-gzip" {
		encoding = "gzip"
	}

	c, ok := codecs[encoding]
	if !ok {
		return nil, "", false
	}

	zr, err := c.reader(bytes.NewReader(data))
	if err != nil {
		return nil, "", false
	}
	defer func() { _ = zr.Close() }()

	decoded, err := io.ReadAll(io.LimitReader(zr, limit))
	if err != nil {
		return nil, "", false
	}

	return decoded, encoding, true
}

// write sends body with status, compressed when the request asked for an encoding.
func write(w http.ResponseWriter, status int, encoding, body string) {
	if encoding == "" {
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)

		return
	}

	w.Header().Set("Content-Encoding", encoding)
	w.Header().Add("Vary", "Accept-Encoding")
	w.WriteHeader(status)

	zw := codecs[encoding].writer(w)
	// The client is the only one who could hear about a failed write.
	_, _ = io.WriteString(zw, body)
	_ = zw.Close()
}
