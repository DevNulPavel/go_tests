package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

var wg sync.WaitGroup
var complete int32

func testFunc(ch chan string, message string) {
	for i := 0; i < 10; i++ {
		//fmt.Println(message)
		//if i%2 == 0 {
		//fmt.Println(message)
		ch <- message
		runtime.Gosched()
		//}
	}
	atomic.AddInt32(&complete, 1)
	wg.Done()
}

func main() {
	ch := make(chan string)

	wg.Add(1)
	go testFunc(ch, "one")
	wg.Add(1)
	go testFunc(ch, "two")

	for {
		fmt.Println(<-ch)
		if atomic.LoadInt32(&complete) == 2 {
			break
		}
	}
	wg.Wait()
}
