// _Channels_ are the pipes that connect concurrent
// goroutines. You can send values into channels from one
// goroutine and receive those values into another
// goroutine.

package main

import "fmt"

func main() {
	messages := make(chan string)
	var i float64 = 0.0

	go func() {
		messages <- "ping"
		fmt.Println("Thread 1 exit")
	}()

	go func() {
		messages <- "ping"
		fmt.Println("Thread 2 exit")
	}()

	go func() {
		messages <- "ping"
		fmt.Println("Thread 3 exit")
	}()

	go func() {
		messages <- "ping"
		fmt.Println("Thread 4 exit")
	}()

	for j := 0; j < 4; j++ {
		msg := <-messages
		i += 1
		fmt.Println(msg)
	}

	fmt.Println(i)

}
