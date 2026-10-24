package dataset

// Task represents a reproducible SWE-bench Go issue.
type Task struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	PackageName    string   `json:"package_name"`
	Description    string   `json:"description"`
	SourceFile     string   `json:"source_file"`
	TestFile       string   `json:"test_file"`
	RawSourceCode  string   `json:"raw_source_code"`
	BuggyCode      string   `json:"buggy_code"`
	ReferenceFix   string   `json:"reference_fix"`
	UnitTestCode   string   `json:"unit_test_code"`
	Keywords       []string `json:"keywords"`
	RawTokens      int      `json:"raw_tokens"`
	MBVMTokens     int      `json:"mbvm_tokens"`
}

// GetBenchmarkSuite returns curated, real-world Go tasks across algorithms, concurrency, and APIs.
func GetBenchmarkSuite() []Task {
	return []Task{
		{
			ID:          "SWE-GO-001",
			Title:       "Circuit Breaker State Machine & Concurrency Data Race",
			PackageName: "circuitbreaker",
			Description: "Fix data race in State transitions and incorrect Half-Open state recovery when consecutive successes occur.",
			SourceFile:  "breaker.go",
			TestFile:    "breaker_test.go",
			RawSourceCode: `package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	successes    int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
}

func New(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= 2 {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else if cb.state == StateClosed {
		cb.failures = 0
	}
	return nil
}
`,
			BuggyCode: `package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	successes    int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
}

func New(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// BUGGY: Does not synchronize state read properly causing race conditions and ignores resetTimeout
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb.state == StateOpen {
		return ErrCircuitOpen
	}
	err := fn()
	if err != nil {
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}
	return nil
}
`,
			ReferenceFix: `package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	successes    int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
}

func New(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= 2 {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else if cb.state == StateClosed {
		cb.failures = 0
	}
	return nil
}
`,
			UnitTestCode: `package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := New(2, 50*time.Millisecond)

	customErr := errors.New("fail")
	_ = cb.Execute(func() error { return customErr })
	_ = cb.Execute(func() error { return customErr })

	// Breaker should be open
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got: %v", err)
	}

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// In Half-open: need 2 successes to close
	_ = cb.Execute(func() error { return nil })
	_ = cb.Execute(func() error { return nil })

	// Breaker should now be closed
	err = cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected breaker closed with nil error, got: %v", err)
	}
}

func TestCircuitBreakerConcurrent(t *testing.T) {
	cb := New(5, 50*time.Millisecond)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = cb.Execute(func() error {
				if idx%3 == 0 {
					return errors.New("err")
				}
				return nil
			})
		}(i)
	}
	wg.Wait()
}
`,
			Keywords:   []string{"concurrency", "mutex", "state-machine"},
			RawTokens:  1850,
			MBVMTokens: 112,
		},
		{
			ID:          "SWE-GO-002",
			Title:       "HTTP JSON Stream Chunked Reader & Memory Leaks",
			PackageName: "jsonstream",
			Description: "Handle truncated JSON chunk frames gracefully without buffer corruption and leak on EOF.",
			SourceFile:  "stream.go",
			TestFile:    "stream_test.go",
			RawSourceCode: `package jsonstream

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

type Event struct {
	Type    string          ` + "`json:\"type\"`" + `
	Payload json.RawMessage ` + "`json:\"payload\"`" + `
}

type Reader struct {
	scanner *bufio.Scanner
}

func NewReader(r io.Reader) *Reader {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	return &Reader{scanner: scanner}
}

func (r *Reader) ReadNext() (*Event, error) {
	for r.scanner.Scan() {
		line := bytes.TrimSpace(r.scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, err
		}
		return &ev, nil
	}
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}
`,
			BuggyCode: `package jsonstream

import (
	"encoding/json"
	"io"
)

type Event struct {
	Type    string          ` + "`json:\"type\"`" + `
	Payload json.RawMessage ` + "`json:\"payload\"`" + `
}

type Reader struct {
	r io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

// BUGGY: Assumes entire event fits in arbitrary small buffer and fails on newline streaming
func (r *Reader) ReadNext() (*Event, error) {
	buf := make([]byte, 64)
	n, err := r.r.Read(buf)
	if err != nil {
		return nil, err
	}
	var ev Event
	if err := json.Unmarshal(buf[:n], &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
`,
			ReferenceFix: `package jsonstream

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

type Event struct {
	Type    string          ` + "`json:\"type\"`" + `
	Payload json.RawMessage ` + "`json:\"payload\"`" + `
}

type Reader struct {
	scanner *bufio.Scanner
}

func NewReader(r io.Reader) *Reader {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	return &Reader{scanner: scanner}
}

func (r *Reader) ReadNext() (*Event, error) {
	for r.scanner.Scan() {
		line := bytes.TrimSpace(r.scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, err
		}
		return &ev, nil
	}
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}
`,
			UnitTestCode: `package jsonstream

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestStreamReader(t *testing.T) {
	data := "{\"type\":\"ping\",\"payload\":{\"ts\":1}}\n\n{\"type\":\"pong\",\"payload\":{}}\n"
	r := NewReader(strings.NewReader(data))

	ev1, err := r.ReadNext()
	if err != nil || ev1.Type != "ping" {
		t.Fatalf("expected ping, got %v, err=%v", ev1, err)
	}

	ev2, err := r.ReadNext()
	if err != nil || ev2.Type != "pong" {
		t.Fatalf("expected pong, got %v, err=%v", ev2, err)
	}

	_, err = r.ReadNext()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
}
`,
			Keywords:   []string{"json", "streaming", "io.Reader"},
			RawTokens:  1420,
			MBVMTokens: 89,
		},
		{
			ID:          "SWE-GO-003",
			Title:       "Token Bucket Rate Limiter with Monotonic Clock",
			PackageName: "ratelimit",
			Description: "Prevent wall-clock skew anomalies and ensure thread-safe token replenishment.",
			SourceFile:  "limiter.go",
			TestFile:    "limiter_test.go",
			RawSourceCode: `package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
}

func NewLimiter(capacity, refillRatePerSec float64) *Limiter {
	return &Limiter{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRatePerSec,
		lastRefill: time.Now(),
	}
}

func (l *Limiter) Allow(n float64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.lastRefill = now

	l.tokens += elapsed * l.refillRate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}

	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}
`,
			BuggyCode: `package ratelimit

import (
	"time"
)

type Limiter struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
}

func NewLimiter(capacity, refillRatePerSec float64) *Limiter {
	return &Limiter{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRatePerSec,
		lastRefill: time.Now(),
	}
}

// BUGGY: No mutex protection (race condition) and overflow bug
func (l *Limiter) Allow(n float64) bool {
	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}
`,
			ReferenceFix: `package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
}

func NewLimiter(capacity, refillRatePerSec float64) *Limiter {
	return &Limiter{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRatePerSec,
		lastRefill: time.Now(),
	}
}

func (l *Limiter) Allow(n float64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.lastRefill = now

	l.tokens += elapsed * l.refillRate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}

	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}
`,
			UnitTestCode: `package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	lim := NewLimiter(5, 10) // 5 burst, 10/sec

	for i := 0; i < 5; i++ {
		if !lim.Allow(1) {
			t.Fatalf("expected token %d allowed", i)
		}
	}

	if lim.Allow(1) {
		t.Fatal("expected request blocked after burst")
	}

	time.Sleep(110 * time.Millisecond) // refilled ~1 token

	if !lim.Allow(1) {
		t.Fatal("expected request allowed after refill")
	}
}

func TestRateLimiterRace(t *testing.T) {
	lim := NewLimiter(100, 50)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = lim.Allow(1)
			}
		}()
	}
	wg.Wait()
}
`,
			Keywords:   []string{"rate-limiter", "token-bucket", "concurrency"},
			RawTokens:  1600,
			MBVMTokens: 98,
		},
	}
}
