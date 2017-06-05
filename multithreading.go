package main

import "fmt"
import "time"
import rnd "math/rand"

func threadFunction1(value int) {
	for i := 0; i < 10; i++ {
		fmt.Println(value, ":", i)

		delay := time.Duration(rnd.Intn(250))
		time.Sleep(time.Millisecond * delay)
	}
}

func sender(message string, channel chan<- string){
	for {
		channel <- message

		delay := time.Duration(rnd.Intn(200))
		time.Sleep(time.Millisecond * delay)
	}
}

func receiver(channel <-chan string) {
	for {
		message := <- channel
		fmt.Println(message)

		//time.Sleep(time.Millisecond * 500)
	}
}

func receiverSelect(channel1 <-chan string, channel2 <-chan string){
	for {
		select{
			case message1 := <- channel1:
				fmt.Println("From chanel1:", message1)
			case message2 := <- channel2:
				fmt.Println("From chanel2:", message2)
			case <- time.After(time.Second):
				fmt.Println("Timeout!")
			/*default:
				fmt.Println("Empty!")*/
		}

		//time.Sleep(time.Millisecond * 500)
	}
}

func main() {
	/*for i := 0; i < 10; i++ {
		go threadFunction1(i)
	}*/

	/*var channel chan string = make(chan string)
	go sender("Ping!", channel)
	go sender("Pong!", channel)
	go receiver(channel)*/

	channel1 := make(chan string, 10)
	channel2 := make(chan string, 10)
	go sender("Ping!", channel1)
	go sender("Pong!", channel2)
	go receiverSelect(channel1, channel2)

	var inputString string
	fmt.Scanln(&inputString)
}