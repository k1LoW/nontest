package nontest_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/k1LoW/nontest"
	nontesting "github.com/k1LoW/nontest/testing"
)

func TestNotTest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	nt := nontest.New(nontest.WithLogger(logger))
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

func TestCleanup(t *testing.T) {
	var got []int
	nt := nontest.New()
	nt.Cleanup(func() {
		got = append(got, 1)
	})
	nt.Cleanup(func() {
		got = append(got, 2)
	})
	nt.Cleanup(func() {
		got = append(got, 3)
	})

	if len(got) != 0 {
		t.Errorf("got %v, want %v", got, []int{})
	}

	nt.TearDown()

	want := []int{3, 2, 1}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestSetenv(t *testing.T) {
	t.Setenv("EXIST_ENV", "exist")
	nt := nontest.New()

	{
		nt.Setenv("EXIST_ENV", "foo")
		got := os.Getenv("EXIST_ENV")
		if want := "foo"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
	{
		nt.Setenv("NOT_EXIST_ENV", "bar")
		got := os.Getenv("NOT_EXIST_ENV")
		if want := "bar"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	nt.TearDown()

	{
		got, ok := os.LookupEnv("EXIST_ENV")
		if want := "exist"; got != want || !ok {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	_, ok := os.LookupEnv("NOT_EXIST_ENV")
	if ok {
		t.Errorf("env NOT_EXIST_ENV is not unset")
	}
}

func TestTempDir(t *testing.T) {
	nt := nontest.New()

	dir := nt.TempDir()

	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Errorf("got %v, want %v", fi, "directory")
	}

	nt.TearDown()

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("temp dir is not removed")
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
