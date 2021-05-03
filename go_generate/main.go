package main

import (
	"log"

	"./queue"
)

func main() {
	q := queue.NewStringQueue()
	q.Push("test")
	testVal := q.Pop()
	log.Print(testVal)
}
