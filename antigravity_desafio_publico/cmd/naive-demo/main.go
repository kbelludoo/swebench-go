package main

import (
    "fmt"
    "os"
)

func main() {
    const trials = 2000
    violations := 0
    for i:=0; i<trials; i++ {
        high,low := make(chan int,1),make(chan int,1)
        high<-1;low<-2
        select {
        case <-high: // correct pick in this trial
        case <-low: violations++ // low picked despite high being ready
        }
    }
    fmt.Printf("Naive select: %d priority violations in %d preloaded trials\n",violations,trials)
    if violations > 0 {
        fmt.Println("EXPECTED FAILURE: ordinary select does NOT guarantee priority")
        os.Exit(1)
    }
    fmt.Println("Anomalous run: no violation observed; retry with more trials")
    os.Exit(2)
}
