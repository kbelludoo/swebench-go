package prioritytrap

import (
    "context"
    "errors"
    "fmt"
    "math/rand"
    "reflect"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

func preloaded(highN, lowN int) (chan int, chan int) {
    high := make(chan int, highN)
    low := make(chan int, lowN)
    for i := 0; i < highN; i++ { high <- i }
    for i := 0; i < lowN; i++ { low <- i }
    close(high)
    close(low)
    return high, low
}

func expected(highN, lowN int) []Event {
    out := make([]Event, 0, highN+lowN)
    for i := 0; i < highN; i++ { out = append(out, Event{High, i}) }
    for i := 0; i < lowN; i++ { out = append(out, Event{Low, i}) }
    return out
}

func collect(ctx context.Context, high, low <-chan int) ([]Event, error) {
    out := make([]Event, 0)
    err := Drain(ctx, high, low, func(e Event) { out = append(out, e) })
    return out, err
}

func TestPublicStrictPriorityPreloaded(t *testing.T) {
    h, l := preloaded(100, 100)
    got, err := collect(context.Background(), h, l)
    if err != nil { t.Fatal(err) }
    want := expected(100, 100)
    if !reflect.DeepEqual(got, want) {
        for i := range got {
            if i >= len(want) || got[i] != want[i] {
                t.Fatalf("priority/FIFO violation at index %d: got=%v expected=%v", i, got[i], want[i])
            }
        }
        t.Fatalf("length mismatch: got=%d want=%d", len(got), len(want))
    }
}

func TestPublicRandomizedPreloaded(t *testing.T) {
    rng := rand.New(rand.NewSource(20261024))
    for trial := 0; trial < 180; trial++ {
        highN, lowN := rng.Intn(120), rng.Intn(120)
        h, l := preloaded(highN, lowN)
        got, err := collect(context.Background(), h, l)
        if err != nil { t.Fatalf("trial=%d: %v", trial, err) }
        if want := expected(highN, lowN); !reflect.DeepEqual(got, want) {
            t.Fatalf("trial=%d, high=%d, low=%d: got=%v want=%v", trial, highN, lowN, got, want)
        }
    }
}

func TestPublicStressTwentyThousandEach(t *testing.T) {
    h, l := preloaded(20_000, 20_000)
    n := 0
    err := Drain(context.Background(), h, l, func(e Event) {
        if n < 20_000 {
            if e.Priority != High || e.Value != n { t.Errorf("high item %d: got=%v", n, e) }
        } else {
            if e.Priority != Low || e.Value != n-20_000 { t.Errorf("low item %d: got=%v", n, e) }
        }
        n++
    })
    if err != nil { t.Fatal(err) }
    if n != 40_000 { t.Fatalf("processed %d, expected 40000", n) }
}

func TestPublicEmptyAndClosedChannels(t *testing.T) {
    cases := []struct{name string; highN, lowN int}{
        {"both_empty", 0, 0}, {"only_high", 10, 0}, {"only_low", 0, 10},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            h,l := preloaded(tc.highN, tc.lowN)
            got, err := collect(context.Background(), h,l)
            if err != nil { t.Fatal(err) }
            if !reflect.DeepEqual(got, expected(tc.highN,tc.lowN)) {t.Fatalf("got=%v",got)}
        })
    }
}

func TestPublicAlreadyCancelledDoesNotDeliver(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    h,l := preloaded(100,100)
    n := 0
    err := Drain(ctx,h,l,func(Event){n++})
    if !errors.Is(err, context.Canceled) {t.Fatalf("got err=%v, want context.Canceled",err)}
    if n != 0 {t.Fatalf("delivered %d events after already-cancelled ctx",n)}
}

func TestPublicCancellationWhileIdle(t *testing.T) {
    ctx,cancel := context.WithCancel(context.Background())
    high,low := make(chan int),make(chan int)
    done := make(chan error,1)
    go func(){done <- Drain(ctx,high,low,func(Event){t.Error("unexpected delivery")})}()
    cancel()
    select {
    case err := <-done:
        if !errors.Is(err, context.Canceled) {t.Fatalf("got=%v",err)}
    case <-time.After(2*time.Second):
        t.Fatal("blocked or leaked after ctx cancellation")
    }
}

func TestPublicLowMustProgressWhenHighIsIdle(t *testing.T) {
    ctx,cancel := context.WithCancel(context.Background())
    defer cancel()
    high := make(chan int) // open but idle: must not block low forever
    low := make(chan int,3)
    for i:=0;i<3;i++ {low<-i}
    close(low)
    received := make(chan Event,3)
    done := make(chan error,1)
    go func(){done<-Drain(ctx,high,low,func(e Event){received<-e})}()
    for i:=0;i<3;i++ {
        select {
        case e:= <-received:
            if e != (Event{Low,i}) {t.Fatalf("low event %d: %v",i,e)}
        case <-time.After(2*time.Second): t.Fatal("low starved by idle high")
        }
    }
    cancel() // high is still open, exit must not wait for its closure
    select {
    case err:= <-done:
        if !errors.Is(err,context.Canceled){t.Fatalf("got=%v",err)}
    case <-time.After(2*time.Second): t.Fatal("did not exit on cancel")
    }
}

func TestPublicAsyncProducer(t *testing.T) {
    high := make(chan int)
    low := make(chan int)
    done := make(chan struct{})
    go func(){
        defer close(done)
        for i:=0;i<250;i++ {high<-i}
        close(high)
        for i:=0;i<250;i++ {low<-i}
        close(low)
    }()
    var got []Event
    ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
    defer cancel()
    if err := Drain(ctx,high,low,func(e Event){got=append(got,e)});err!=nil {t.Fatal(err)}
    <-done
    // Because phases do not overlap, order is deterministic.
    if want:=expected(250,250);!reflect.DeepEqual(got,want){t.Fatalf("unexpected staged order: got=%v",got)}
}

func TestPublicSerialCallbacks(t *testing.T) {
    h,l:=preloaded(200,200)
    var active,maxActive atomic.Int64
    err:=Drain(context.Background(),h,l,func(Event){
        count:=active.Add(1)
        for {
            prior:=maxActive.Load()
            if count<=prior || maxActive.CompareAndSwap(prior,count){break}
        }
        time.Sleep(100*time.Microsecond)
        active.Add(-1)
    })
    if err!=nil{t.Fatal(err)}
    if maxActive.Load()!=1{t.Fatalf("callbacks ran concurrently: max=%d",maxActive.Load())}
}

func TestPublicIndependentConcurrentRunsRace(t *testing.T) {
    const workers=12
    var wg sync.WaitGroup
    errs:=make(chan error,workers)
    for w:=0;w<workers;w++ {
        wg.Add(1)
        go func(id int){
            defer wg.Done()
            h,l:=preloaded(120,80)
            got,err:=collect(context.Background(),h,l)
            if err!=nil{errs<-err;return}
            if !reflect.DeepEqual(got,expected(120,80)) {errs<-fmt.Errorf("worker %d wrong ordering",id)}
        }(w)
    }
    wg.Wait()
    close(errs)
    for err:=range errs {t.Error(err)}
}
