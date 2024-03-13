package nontest

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	nontesting "github.com/k1LoW/nontest/testing"
)

func TestNotTest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	nt := New(WithLogger(logger))
	ts := testServer(nt)
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code is not 200: %d", res.StatusCode)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if string(b) != "OK" {
		t.Errorf("response body is not 'OK': %s", string(b))
	}
	{
		want := `"level":"INFO"`
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("log does not contain %q: %s", want, buf.String())
		}
	}
	{
		want := `"msg":"Server started"`
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("log does not contain %q: %s", want, buf.String())
		}
	}
	nt.TearDown()
}

func TestTestServer(t *testing.T) {
	ts := testServer(t)
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code is not 200: %d", res.StatusCode)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if string(b) != "OK" {
		t.Errorf("response body is not 'OK': %s", string(b))
	}
}

func testServer(t nontesting.TB) *httptest.Server {
	t.Helper()
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	ts := httptest.NewServer(h)
	t.Log("Server started")
	t.Cleanup(func() {
		ts.Close()
	})
	return ts
}
