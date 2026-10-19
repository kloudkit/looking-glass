package reflect

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

const (
	EncodingHeader = "X-Glass-Encoding"
	EncodingQuery  = "glass-encoding"
)

// decompress decodes a gzip or deflate body, keeping at most limit bytes so a small upload
// can't expand past the body limit. Any other encoding, or a stream that doesn't decode, is
// left as it arrived.
func decompress(encoding string, data []byte, limit int64) ([]byte, string, bool) {
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	var (
		zr  io.ReadCloser
		err error
	)
	switch encoding {
	case "gzip", "x-gzip":
		zr, err = gzip.NewReader(bytes.NewReader(data))
	case "deflate":
		zr, err = zlib.NewReader(bytes.NewReader(data))
	default:
		return nil, "", false
	}
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

func validEncoding(encoding string) bool {
	return encoding == "gzip" || encoding == "deflate"
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

	var zw io.WriteCloser = gzip.NewWriter(w)
	if encoding == "deflate" {
		zw = zlib.NewWriter(w)
	}
	// The client is the only one who could hear about a failed write.
	_, _ = io.WriteString(zw, body)
	_ = zw.Close()
}
