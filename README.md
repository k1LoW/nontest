# nontest

Use implementations created as test helpers for nontest purposes.

## Overview

nontest provides an implementation compatible with Go's standard `testing.TB` interface, allowing you to write helper functions once and use them in both test code and production code.

## Features

- **Full `testing.TB` compatibility** - Implements Go's standard testing.TB interface
- **slog-based logging** - Outputs structured logs in JSON format
- **Automatic cleanup** - FILO execution of Cleanup functions, automatic restoration of environment variables, and automatic deletion of temporary directories
- **Thread-safe** - Concurrent access control with sync.Mutex
- **Flexible behavior control** - Customizable exit behavior with AllowExit option
- **Zero dependencies** - Uses only Go standard library

## Installation

```bash
go get github.com/k1LoW/nontest
```

## Usage

### Basic Example

```go
package main

import (
    "log/slog"
    "net/http"
    "net/http/httptest"
    
    "github.com/k1LoW/nontest"
    "github.com/k1LoW/nontest/testing"
)

// Helper function usable in both test and production code
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

// Use in actual tests
func TestTestServer(t *testing.T) {
    ts := testServer(t)  // Pass standard testing.T
    // ... test logic
}

// Use in production/non-test code
func main() {
    logger := slog.Default()
    nt := nontest.New(nontest.WithLogger(logger))
    ts := testServer(nt)  // Pass nontest instance
    // ... production logic
    nt.TearDown()  // Explicit cleanup
}
```

### API

#### Creating a nontest instance

```go
nt := nontest.New(opts ...Option)
```

#### Options

- `WithLogger(logger *slog.Logger)` - Sets a custom logger
- `AllowExit()` - Allows runtime.Goexit() on FailNow/SkipNow

## Use Cases

- Reuse test helper functions in CLI tools
- Utilize convenient test features (TempDir, Setenv, etc.) in production code
- Create helper functions with a common interface for tests and production
- Leverage test infrastructure in debug or demo environments
