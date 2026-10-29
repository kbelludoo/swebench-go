package prioritytrap

import (
	"context"
)

// Priority identifies which input delivered an event.
type Priority uint8

const (
	High Priority = iota + 1
	Low
)

// Event is delivered synchronously to emit.
type Event struct {
	Priority Priority
	Value    int
}

// Drain satisfies the contract in README.md. Implement it from scratch.
// The input channels may be closed or nil; emit is non-nil.
func Drain(ctx context.Context, high <-chan int, low <-chan int, emit func(Event)) error {
	var pendingLow int
	var hasPendingLow bool

	for {
		// 1. Immediate cancellation check
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 2. Termination check: if both sources are inactive and no low item is pending
		if high == nil && low == nil && !hasPendingLow {
			return nil
		}

		// 3. Strict priority check: always exhaust high if a message is available
		if high != nil {
			select {
			case val, ok := <-high:
				if !ok {
					high = nil
					continue
				}
				emit(Event{Priority: High, Value: val})
				continue
			default:
				// High has no message ready right now
			}
		}

		// 4. If we were holding a pending low item and high had nothing, emit it now
		if hasPendingLow {
			emit(Event{Priority: Low, Value: pendingLow})
			hasPendingLow = false
			continue
		}

		// 5. Non-blocking check on low if available
		if low != nil {
			select {
			case val, ok := <-low:
				if !ok {
					low = nil
					continue
				}
				if high == nil {
					emit(Event{Priority: Low, Value: val})
				} else {
					pendingLow = val
					hasPendingLow = true
				}
				continue
			default:
				// Low also has no message ready right now
			}
		}

		// 6. Both channels are currently idle. Block until an event occurs (no busy-spin).
		select {
		case <-ctx.Done():
			return ctx.Err()
		case val, ok := <-high:
			if !ok {
				high = nil
				continue
			}
			emit(Event{Priority: High, Value: val})
		case val, ok := <-low:
			if !ok {
				low = nil
				continue
			}
			if high == nil {
				emit(Event{Priority: Low, Value: val})
			} else {
				pendingLow = val
				hasPendingLow = true
			}
		}
	}
}
