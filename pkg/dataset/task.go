package dataset

// Task represents a reproducible SWE-bench Go issue.
type Task struct {
	ID                      string   `json:"id"`
	Title                   string   `json:"title"`
	PackageName             string   `json:"package_name"`
	Description             string   `json:"description"`
	SourceFile              string   `json:"source_file"`
	TestFile                string   `json:"test_file"`
	TargetSymbol            string   `json:"target_symbol"`
	RawSourceCode           string   `json:"raw_source_code"`
	BuggyCode               string   `json:"buggy_code"`
	StandardAIPatch         string   `json:"standard_ai_patch"`
	StandardAIFailureReason string   `json:"standard_ai_failure_reason"`
	ReferenceFix            string   `json:"reference_fix"`
	UnitTestCode            string   `json:"unit_test_code"`
	Keywords                []string `json:"keywords"`
	RawTokens               int      `json:"raw_tokens"`
	MBVMTokens              int      `json:"mbvm_tokens"`
	MBVMDBTokens            int      `json:"mbvmdb_tokens"`
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
			StandardAIPatch: `package circuitbreaker

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

// Standard AI Naive Fix: Ignores HalfOpen state machine transition and omits reset logic, failing recovery.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) <= cb.resetTimeout {
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
	return nil
}
`,
			StandardAIFailureReason: "State machine contract violation: omits Half-Open consecutive success transition and leaves breaker in invalid recovery state.",
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

	// In Half-open: 1st success should maintain HalfOpen state
	_ = cb.Execute(func() error { return nil })
	if cb.state != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen after 1 success, got: %v", cb.state)
	}

	// 2nd success closes the breaker
	_ = cb.Execute(func() error { return nil })
	if cb.state != StateClosed {
		t.Fatalf("expected StateClosed after 2 successes, got: %v", cb.state)
	}

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
			Keywords:     []string{"concurrency", "mutex", "state-machine"},
			TargetSymbol: "Execute",
			RawTokens:    1850,
			MBVMTokens:   112,
			MBVMDBTokens: 69,
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
			StandardAIPatch: `package jsonstream

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

// Standard AI Naive Fix: Reads whole unbuffered stream into static slice, corrupting multiple JSON frame sequences.
func (r *Reader) ReadNext() (*Event, error) {
	buf := make([]byte, 1024)
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
			StandardAIFailureReason: "Stream framing failure: unbuffered raw read consumes multiple newline-delimited frames into a single payload, crashing json.Unmarshal.",
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
			Keywords:     []string{"json", "streaming", "io.Reader"},
			TargetSymbol: "ReadNext",
			RawTokens:    1420,
			MBVMTokens:   89,
			MBVMDBTokens: 37,
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
			StandardAIPatch: `package ratelimit

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

// Standard AI Naive Fix: Omits sync.Mutex concurrency protection and monotonic elapsed refill math.
func (l *Limiter) Allow(n float64) bool {
	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}
`,
			StandardAIFailureReason: "Concurrency & time math omission: missing sync.Mutex causing race conditions and omitting monotonic elapsed token replenishment.",
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
			Keywords:     []string{"rate-limiter", "token-bucket", "concurrency"},
			TargetSymbol: "Allow",
			RawTokens:    1600,
			MBVMTokens:   98,
			MBVMDBTokens: 53,
		},
		{
			ID:          "SWE-GO-004",
			Title:       "gin-gonic/gin #3017: Context.Copy() Concurrency Map Race & Panic",
			PackageName: "gincontext",
			Description: "Context.Copy() shallow copies Keys map without deep clone, causing concurrent map read and map write panic when background goroutines access Keys.",
			SourceFile:  "context.go",
			TestFile:    "context_test.go",
			RawSourceCode: `package gincontext

import (
	"sync"
)

type Context struct {
	mu   sync.RWMutex
	Keys map[string]any
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Keys[key]
	return val, ok
}

func (c *Context) Copy() *Context {
	cp := &Context{}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Keys != nil {
		cp.Keys = make(map[string]any, len(c.Keys))
		for k, v := range c.Keys {
			cp.Keys[k] = v
		}
	}
	return cp
}
`,
			BuggyCode: `package gincontext

import (
	"sync"
)

type Context struct {
	mu   sync.RWMutex
	Keys map[string]any
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Keys[key]
	return val, ok
}

// BUGGY (gin-gonic/gin #3017): Shallow copies reference to map, causing fatal concurrent map read/write
func (c *Context) Copy() *Context {
	return &Context{
		Keys: c.Keys,
	}
}
`,
			StandardAIPatch: `package gincontext

import (
	"sync"
)

type Context struct {
	mu   sync.RWMutex
	Keys map[string]any
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Keys[key]
	return val, ok
}

// Standard AI Naive Attempt: Allocates new struct but still shallow assigns the shared map pointer
func (c *Context) Copy() *Context {
	cp := &Context{
		Keys: c.Keys,
	}
	return cp
}
`,
			StandardAIFailureReason: "Third-Party gin-gonic/gin #3017 failure: shallow copy of map reference causes fatal concurrent map read and map write under goroutine workload.",
			ReferenceFix: `package gincontext

import (
	"sync"
)

type Context struct {
	mu   sync.RWMutex
	Keys map[string]any
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Keys[key]
	return val, ok
}

func (c *Context) Copy() *Context {
	cp := &Context{}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Keys != nil {
		cp.Keys = make(map[string]any, len(c.Keys))
		for k, v := range c.Keys {
			cp.Keys[k] = v
		}
	}
	return cp
}
`,
			UnitTestCode: `package gincontext

import (
	"fmt"
	"sync"
	"testing"
)

func TestContextConcurrentCopy(t *testing.T) {
	c := &Context{}
	c.Set("initial", "val")

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cp := c.Copy()
			cp.Set(fmt.Sprintf("k_%d", id), id)
			c.Set(fmt.Sprintf("orig_%d", id), id)
			_, _ = cp.Get(fmt.Sprintf("k_%d", id))
			_, _ = c.Get("initial")
		}(i)
	}
	wg.Wait()
}
`,
			Keywords:     []string{"gin-gonic", "context", "deep-copy", "concurrency"},
			TargetSymbol: "Copy",
			RawTokens:    1520,
			MBVMTokens:   84,
			MBVMDBTokens: 41,
		},
		{
			ID:          "SWE-GO-005",
			Title:       "etcd-io/etcd #13421: Watcher Goroutine Leak Verified with Uber Goleak",
			PackageName: "etcdwatcher",
			Description: "Watcher background pump goroutine fails to listen to context cancellation on channel send, causing goroutines to leak forever on cancelled stream.",
			SourceFile:  "watcher.go",
			TestFile:    "watcher_test.go",
			RawSourceCode: `package etcdwatcher

import (
	"context"
	"sync"
)

type Event struct {
	Key   string
	Value string
}

type Watcher struct {
	mu     sync.Mutex
	stopCh chan struct{}
	closed bool
}

func NewWatcher() *Watcher {
	return &Watcher{
		stopCh: make(chan struct{}),
	}
}

func (w *Watcher) Watch(ctx context.Context, in <-chan Event, out chan<- Event) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case ev, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- ev:
				case <-ctx.Done():
					return
				case <-w.stopCh:
					return
				}
			}
		}
	}()
}

func (w *Watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		w.closed = true
		close(w.stopCh)
	}
}
`,
			BuggyCode: `package etcdwatcher

import (
	"context"
	"sync"
)

type Event struct {
	Key   string
	Value string
}

type Watcher struct {
	mu     sync.Mutex
	stopCh chan struct{}
	closed bool
}

func NewWatcher() *Watcher {
	return &Watcher{
		stopCh: make(chan struct{}),
	}
}

// BUGGY: Blocks indefinitely trying to send on 'out' without listening to context cancellation
func (w *Watcher) Watch(ctx context.Context, in <-chan Event, out chan<- Event) {
	go func() {
		for ev := range in {
			out <- ev // Leaks goroutine when receiver abandons 'out' channel
		}
	}()
}

func (w *Watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		w.closed = true
		close(w.stopCh)
	}
}
`,
			StandardAIPatch: `package etcdwatcher

import (
	"context"
	"sync"
)

type Event struct {
	Key   string
	Value string
}

type Watcher struct {
	mu     sync.Mutex
	stopCh chan struct{}
	closed bool
}

func NewWatcher() *Watcher {
	return &Watcher{
		stopCh: make(chan struct{}),
	}
}

// Standard AI Naive Attempt: Checks ctx.Done on read, but blocks naked on write to 'out', leaking goroutine
func (w *Watcher) Watch(ctx context.Context, in <-chan Event, out chan<- Event) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-in:
				if !ok {
					return
				}
				out <- ev // BUG: Blocks forever if consumer stopped listening, caught by Uber Goleak
			}
		}
	}()
}

func (w *Watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		w.closed = true
		close(w.stopCh)
	}
}
`,
			StandardAIFailureReason: "Third-Party etcd #13421 / Uber Goleak failure: unbuffered send blocks indefinitely on cancelled context, detected by go.uber.org/goleak.",
			ReferenceFix: `package etcdwatcher

import (
	"context"
	"sync"
)

type Event struct {
	Key   string
	Value string
}

type Watcher struct {
	mu     sync.Mutex
	stopCh chan struct{}
	closed bool
}

func NewWatcher() *Watcher {
	return &Watcher{
		stopCh: make(chan struct{}),
	}
}

func (w *Watcher) Watch(ctx context.Context, in <-chan Event, out chan<- Event) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case ev, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- ev:
				case <-ctx.Done():
					return
				case <-w.stopCh:
					return
				}
			}
		}
	}()
}

func (w *Watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		w.closed = true
		close(w.stopCh)
	}
}
`,
			UnitTestCode: `package etcdwatcher

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestWatcherGoroutineLeakWithUberGoleak(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(context.Background())
	w := NewWatcher()

	in := make(chan Event)
	out := make(chan Event)

	w.Watch(ctx, in, out)

	// Send an event that is not drained by reader
	in <- Event{Key: "foo", Value: "bar"}

	// Cancel context and close watcher
	cancel()
	w.Close()
	close(in)

	// Allow goroutines a brief moment to react to cancellation
	time.Sleep(50 * time.Millisecond)
}
`,
			Keywords:     []string{"etcd", "goleak", "goroutine-leak", "uber"},
			TargetSymbol: "Watch",
			RawTokens:    1750,
			MBVMTokens:   95,
			MBVMDBTokens: 46,
		},
	}
}
