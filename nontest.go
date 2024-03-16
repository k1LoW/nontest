package nontest

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/k1LoW/nontest/testing"
)

var _ testing.TB = (*nonTest)(nil)

type nonTest struct {
	logger       *slog.Logger
	allowExit    bool
	failed       bool
	skipped      bool
	finished     bool
	prevEnvs     map[string]*string
	tmpDirs      []string
	cleanupFuncs []func()
	mu           sync.Mutex
}

type Option func(*nonTest)

// WithLogger returns an Option to set logger.
func WithLogger(logger *slog.Logger) Option {
	return func(t *nonTest) {
		t.logger = logger
	}
}

// AllowExit returns an Option to allow exit.
func AllowExit() Option {
	return func(t *nonTest) {
		t.allowExit = true
	}
}

// New returns a new nonTest.
func New(opts ...Option) *nonTest {
	t := &nonTest{
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		prevEnvs: map[string]*string{},
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *nonTest) Cleanup(f func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cleanupFuncs = append(t.cleanupFuncs, f)
}

func (t *nonTest) TearDown() {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Call cleanupFuncs FILO
	for i := len(t.cleanupFuncs) - 1; i >= 0; i-- {
		t.cleanupFuncs[i]()
	}
	// Revert envs
	for k, v := range t.prevEnvs {
		if v != nil {
			os.Setenv(k, *v)
		} else {
			os.Unsetenv(k)
		}
	}
	// Remove tmpDirs
	for _, dir := range t.tmpDirs {
		os.RemoveAll(dir)
	}
}

func (t *nonTest) Error(args ...any) {
	t.logger.Error(fmt.Sprintln(args...))
	t.Fail()
}

func (t *nonTest) Errorf(format string, args ...any) {
	t.logger.Error(fmt.Sprintf(format, args...))
	t.Fail()
}

func (t *nonTest) Fail() {
	t.failed = true
}

func (t *nonTest) FailNow() {
	t.Fail()
	t.finished = true
	t.TearDown()
	if t.allowExit {
		runtime.Goexit()
	}
}

func (t *nonTest) Failed() bool {
	return t.failed
}

func (t *nonTest) Fatal(args ...any) {
	t.logger.Error(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
	t.FailNow()
}

func (t *nonTest) Fatalf(format string, args ...any) {
	t.logger.Error(strings.TrimSuffix(fmt.Sprintf(format, args...), "\n"))
	t.FailNow()
}

func (t *nonTest) Helper() {}

func (t *nonTest) Log(args ...interface{}) {
	t.logger.Info(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func (t *nonTest) Logf(format string, args ...any) {
	t.logger.Info(strings.TrimSuffix(fmt.Sprintf(format, args...), "\n"))
}

func (t *nonTest) Name() string {
	return ""
}

func (t *nonTest) Setenv(key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.prevEnvs[key]
	if ok {
		os.Setenv(key, value)
		return
	}
	if prev, ok := os.LookupEnv(key); ok {
		t.prevEnvs[key] = &prev
	} else {
		t.prevEnvs[key] = nil
	}
	os.Setenv(key, value)
}

func (t *nonTest) Skip(args ...any) {
	t.logger.Info(strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
	t.SkipNow()
}

func (t *nonTest) SkipNow() {
	t.skipped = true
	t.finished = true
	t.TearDown()
	if t.allowExit {
		runtime.Goexit()
	}
}

func (t *nonTest) Skipf(format string, args ...any) {
	t.logger.Info(strings.TrimSuffix(fmt.Sprintf(format, args...), "\n"))
	t.SkipNow()
}

func (t *nonTest) Skipped() bool {
	return t.skipped
}

func (t *nonTest) TempDir() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	dir, _ := os.MkdirTemp("", "tmp")
	t.tmpDirs = append(t.tmpDirs, dir)
	return dir
}
