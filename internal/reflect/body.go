package reflect

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"unicode/utf8"
)

type upload struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Bytes       int64  `json:"bytes"`
}

func (ref *reflection) readBody(r *http.Request, maxBody int64) {
	ref.BodyEncoding = "utf-8"
	if r.Body == nil {
		return
	}

	data, _ := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	ref.Truncated = int64(len(data)) > maxBody
	if ref.Truncated {
		data = trimPartialRune(data[:maxBody])
	}

	// A cut-off body still reports its full size when the client declared one.
	ref.BodyBytes = int64(len(data))
	if ref.Truncated && r.ContentLength > ref.BodyBytes {
		ref.BodyBytes = r.ContentLength
	}

	ref.Body = string(data)
	if !utf8.Valid(data) {
		ref.Body, ref.BodyEncoding = base64.StdEncoding.EncodeToString(data), "base64"
	}

	// Only a whole body can be parsed.
	if len(data) == 0 || ref.Truncated {
		return
	}

	if json.Valid(data) {
		ref.JSON = data
	}

	ref.Form, ref.Files = parseForm(r.Header.Get("Content-Type"), data)
}

// trimPartialRune drops a character the cut-off split, so a text body stays valid UTF-8.
func trimPartialRune(data []byte) []byte {
	for i := 1; i < utf8.UTFMax && i <= len(data); i++ {
		if utf8.RuneStart(data[len(data)-i]) {
			if !utf8.FullRune(data[len(data)-i:]) {
				return data[:len(data)-i]
			}

			break
		}
	}

	return data
}

func parseForm(contentType string, data []byte) (map[string][]string, map[string][]upload) {
	mediaType, params, _ := mime.ParseMediaType(contentType)

	switch mediaType {
	case "application/x-www-form-urlencoded":
		if form, err := url.ParseQuery(string(data)); err == nil {
			return form, nil
		}
	case "multipart/form-data":
		return parseMultipart(data, params["boundary"])
	}

	return nil, nil
}

// parseMultipart lists the form's fields, and its files by name and size; their contents are in the body.
func parseMultipart(data []byte, boundary string) (map[string][]string, map[string][]upload) {
	form, err := multipart.NewReader(bytes.NewReader(data), boundary).ReadForm(int64(len(data)))
	if err != nil {
		return nil, nil
	}
	defer func() { _ = form.RemoveAll() }()

	files := map[string][]upload{}
	for name, headers := range form.File {
		for _, h := range headers {
			files[name] = append(files[name], upload{
				Filename:    h.Filename,
				ContentType: h.Header.Get("Content-Type"),
				Bytes:       h.Size,
			})
		}
	}

	return form.Value, files
}
