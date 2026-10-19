package reflect

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyParsesJSON(t *testing.T) {
	ref := reflectOf("POST", "/x", `{"name":"ada"}`)

	if string(ref.JSON) != `{"name":"ada"}` {
		t.Errorf("json = %s", ref.JSON)
	}

	if out := ref.json(); !strings.Contains(out, `"json": {`) || !strings.Contains(out, `"name": "ada"`) {
		t.Errorf("expected the body nested as JSON:\n%s", out)
	}
}

func TestBodyLeavesTextAlone(t *testing.T) {
	ref := reflectOf("POST", "/x", "not json")

	if ref.JSON != nil || ref.Form != nil || ref.Files != nil {
		t.Errorf("plain text parsed as json=%s form=%v files=%v", ref.JSON, ref.Form, ref.Files)
	}

	if out := ref.json(); !strings.Contains(out, `"json": null`) {
		t.Errorf("expected json null:\n%s", out)
	}
}

func TestBodyBase64EncodesBinary(t *testing.T) {
	binary := "\xff\xfe\x00png"
	ref := reflectOf("POST", "/x", binary)

	if ref.BodyEncoding != "base64" || ref.Body != base64.StdEncoding.EncodeToString([]byte(binary)) {
		t.Errorf("body = %q (%s)", ref.Body, ref.BodyEncoding)
	}

	if ref.BodyBytes != int64(len(binary)) {
		t.Errorf("bodyBytes = %d, want the decoded size %d", ref.BodyBytes, len(binary))
	}
}

func TestBodyParsesURLEncodedForm(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader("name=ada&tag=a&tag=b"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ref := build(req, opts(1024))

	if got := ref.Form["tag"]; len(got) != 2 || got[1] != "b" || ref.Form["name"][0] != "ada" {
		t.Errorf("form = %v", ref.Form)
	}
}

func TestBodyParsesMultipart(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("name", "ada"); err != nil {
		t.Fatal(err)
	}
	fw, err := mw.CreateFormFile("avatar", "ada.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("\x89PNG....")); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/x", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	ref := build(req, opts(1<<20))

	if ref.Form["name"][0] != "ada" {
		t.Errorf("form = %v", ref.Form)
	}

	files := ref.Files["avatar"]
	if len(files) != 1 || files[0].Filename != "ada.png" || files[0].Bytes != 8 ||
		files[0].ContentType != "application/octet-stream" {
		t.Errorf("files = %+v", ref.Files)
	}
}

func TestBodyTruncatedIsNotParsed(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader(`{"name":"ada"}`))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ref := build(req, opts(5))

	if ref.JSON != nil || ref.Form != nil {
		t.Errorf("a cut-off body was parsed: json=%s form=%v", ref.JSON, ref.Form)
	}
}

func TestBodyTruncatedKeepsUTF8(t *testing.T) {
	// "é" is two bytes; a cut after 4 bytes splits the second one.
	req := httptest.NewRequest("POST", "/x", strings.NewReader("aéééé"))

	ref := build(req, opts(4))

	if ref.BodyEncoding != "utf-8" || ref.Body != "aé" {
		t.Errorf("body = %q (%s), want %q", ref.Body, ref.BodyEncoding, "aé")
	}
}

func TestBodyTruncatedKeepsDeclaredSize(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader(strings.Repeat("x", 50)))

	ref := build(req, opts(10))

	if !ref.Truncated || len(ref.Body) != 10 || ref.BodyBytes != 50 {
		t.Errorf("body %d bytes, bodyBytes %d; want 10 and 50", len(ref.Body), ref.BodyBytes)
	}
}
