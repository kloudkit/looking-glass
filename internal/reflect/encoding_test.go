package reflect

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func compressed(t *testing.T, encoding, s string) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	var zw io.WriteCloser = gzip.NewWriter(&buf)
	if encoding == "deflate" {
		zw = zlib.NewWriter(&buf)
	}
	if _, err := io.WriteString(zw, s); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	return &buf
}

func TestBodyDecompresses(t *testing.T) {
	for _, encoding := range []string{"gzip", "deflate"} {
		req := httptest.NewRequest("POST", "/x", compressed(t, encoding, `{"name":"ada"}`))
		req.Header.Set("Content-Encoding", encoding)

		ref := build(req, testOpts)

		if ref.Body != `{"name":"ada"}` || ref.ContentEncoding != encoding || ref.JSON == nil {
			t.Errorf("%s: body = %q (from %q), json = %s", encoding, ref.Body, ref.ContentEncoding, ref.JSON)
		}
	}
}

func TestBodyDecompressionIsBounded(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", compressed(t, "gzip", strings.Repeat("x", 1<<20)))
	req.Header.Set("Content-Encoding", "gzip")

	ref := build(req, opts(100))

	if !ref.Truncated || len(ref.Body) != 100 || ref.BodyBytes != 100 {
		t.Errorf("body %d bytes, bodyBytes %d, truncated %v; want 100, 100, true", len(ref.Body), ref.BodyBytes, ref.Truncated)
	}
}

func TestBodyBadCompressionStaysRaw(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")

	ref := build(req, testOpts)

	if ref.Body != "not gzip" || ref.ContentEncoding != "" {
		t.Errorf("body = %q (from %q)", ref.Body, ref.ContentEncoding)
	}
}

func TestHandlerCompressesOnRequest(t *testing.T) {
	for _, encoding := range []string{"gzip", "deflate"} {
		req := httptest.NewRequest("GET", "/x?glass-encoding="+encoding, nil)
		req.Header.Set(formatOption.header, "json")

		rec := serve(req, testOpts)

		if got := rec.Header().Get("Content-Encoding"); got != encoding {
			t.Fatalf("%s: content-encoding = %q", encoding, got)
		}

		var zr io.Reader
		var err error
		if encoding == "gzip" {
			zr, err = gzip.NewReader(rec.Body)
		} else {
			zr, err = zlib.NewReader(rec.Body)
		}
		if err != nil {
			t.Fatal(err)
		}

		out, _ := io.ReadAll(zr)
		if !strings.Contains(string(out), `"method": "GET"`) {
			t.Errorf("%s: decoded response = %q", encoding, out)
		}
	}
}

func TestHandlerRejectsUnknownEncoding(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set(encodingOption.header, "br")

	rec := serve(req, testOpts)

	if rec.Code != 400 || !strings.Contains(rec.Body.String(), encodingOption.header) {
		t.Errorf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestHandlerDoesNotCompressByDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := serve(req, testOpts)

	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Errorf("content-encoding = %q, want none", got)
	}
}
