//go:build goleak

package prioritytrap

import (
    "context"
    "errors"
    "testing"
    "time"

    "go.uber.org/goleak"
)

func TestHoldoutGoleakMixedExitPaths(t *testing.T) {
    defer goleak.VerifyNone(t)
    for i:=0;i<150;i++ {
        if i%2==0 {
            high,low:=preloaded(20,30)
            if _,err:=collect(context.Background(),high,low);err!=nil{t.Fatal(err)}
        } else {
            ctx,cancel:=context.WithCancel(context.Background())
            high,low:=make(chan int),make(chan int)
            done:=make(chan error,1)
            go func(){done<-Drain(ctx,high,low,func(Event){})}()
            cancel()
            select {
            case err:= <-done:
                if !errors.Is(err,context.Canceled){t.Fatal(err)}
            case <-time.After(time.Second):t.Fatal("goroutine did not exit")
            }
        }
    }
}
