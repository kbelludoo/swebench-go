package prioritytrap

import (
    "context"
    "errors"
    "reflect"
    "testing"
    "time"
)

// HOLDOUT: Inject high-priority work during a synchronous callback, when low is queued.
func TestHoldoutHighArrivalInLowHandler(t *testing.T) {
    high:=make(chan int,10)
    low:=make(chan int,3)
    for i:=0;i<3;i++ { low<-i }
    close(low)
    var got []Event
    err:=Drain(context.Background(),high,low,func(e Event){
        got=append(got,e)
        if len(got)==1 {high<-900;high<-901;close(high)}
    })
    if err!=nil{t.Fatal(err)}
    want:=[]Event{{Low,0},{High,900},{High,901},{Low,1},{Low,2}}
    if !reflect.DeepEqual(got,want){t.Fatalf("new high backlog not prioritized: got=%v want=%v",got,want)}
}

func TestHoldoutRefillHighWhileLowQueued(t *testing.T) {
    high:=make(chan int,128)
    low:=make(chan int,128)
    high<-0
    for i:=0;i<50;i++ {low<-i}
    close(low)
    got:=make([]Event,0,151)
    err:=Drain(context.Background(),high,low,func(e Event){
        got=append(got,e)
        if e.Priority==High {
            if e.Value<100 {high<-e.Value+1} else {close(high)}
        }
    })
    if err!=nil{t.Fatal(err)}
    if want:=expected(101,50);!reflect.DeepEqual(got,want){t.Fatalf("refill priority violated: got=%v",got)}
}

func TestHoldoutNilInputs(t *testing.T) {
    got,err:=collect(context.Background(),nil,nil)
    if err!=nil{t.Fatal(err)}
    if len(got)!=0{t.Fatalf("nil/nil delivered=%v",got)}
    low:=make(chan int,2);low<-7;low<-8;close(low)
    events,err:=collect(context.Background(),nil,low)
    if err!=nil{t.Fatal(err)}
    if want:=[]Event{{Low,7},{Low,8}};!reflect.DeepEqual(events,want){t.Fatalf("got=%v want=%v",events,want)}
    high:=make(chan int,1);high<-9;close(high)
    events,err=collect(context.Background(),high,nil)
    if err!=nil{t.Fatal(err)}
    if want:=[]Event{{High,9}};!reflect.DeepEqual(events,want){t.Fatalf("got=%v want=%v",events,want)}
}

func TestHoldoutCancellationFromCallbackWhileBacklogged(t *testing.T) {
    ctx,cancel:=context.WithCancel(context.Background())
    defer cancel()
    high,low:=preloaded(20_000,20_000)
    count:=0
    done:=make(chan error,1)
    go func(){done<-Drain(ctx,high,low,func(Event){count++; if count==1{cancel()}})}()
    select {
    case err:= <-done:
        if !errors.Is(err,context.Canceled){t.Fatalf("got=%v",err)}
        if count>=100 {t.Fatalf("processed %d items after callback cancelled ctx",count)}
    case <-time.After(2*time.Second):t.Fatal("cancellation delayed or ignored")
    }
}

func TestHoldoutWaitForLateProducerWithoutSpinConstraint(t *testing.T) {
    ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
    defer cancel()
    high,low:=make(chan int),make(chan int)
    produced:=make(chan struct{})
    go func(){
        defer close(produced)
        time.Sleep(25*time.Millisecond)
        high<-1
        close(high)
        low<-2
        close(low)
    }()
    got,err:=collect(ctx,high,low)
    if err!=nil{t.Fatal(err)}
    <-produced
    if want:=[]Event{{High,1},{Low,2}};!reflect.DeepEqual(got,want){t.Fatalf("got=%v want=%v",got,want)}
}

func TestHoldoutManyAlreadyCancelledNoCallbacks(t *testing.T) {
    for i:=0;i<500;i++ {
        ctx,cancel:=context.WithCancel(context.Background());cancel()
        h,l:=preloaded(5,5)
        n:=0
        err:=Drain(ctx,h,l,func(Event){n++})
        if !errors.Is(err,context.Canceled) || n!=0 {t.Fatalf("round=%d n=%d err=%v",i,n,err)}
    }
}
