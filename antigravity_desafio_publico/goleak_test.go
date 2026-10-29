//go:build goleak

package prioritytrap

import (
    "context"
    "errors"
    "testing"
    "time"

    "go.uber.org/goleak"
)

func TestPublicUberGoleakAfterCancellation(t *testing.T) {
    defer goleak.VerifyNone(t)
    for i:=0; i<100; i++ {
        ctx,cancel := context.WithCancel(context.Background())
        high,low := make(chan int),make(chan int)
        done := make(chan error,1)
        go func(){done<-Drain(ctx,high,low,func(Event){})}()
        cancel()
        select {
        case err:= <-done:
            if !errors.Is(err,context.Canceled){t.Fatalf("round=%d err=%v",i,err)}
        case <-time.After(2*time.Second):t.Fatalf("round=%d goroutine stuck",i)
        }
    }
}
