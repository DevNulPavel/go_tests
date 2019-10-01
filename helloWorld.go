package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

type Point struct {
	x float32
	y float32
}

func (p *Point) length() float32 {
	return float32(math.Sqrt(float64(p.x*p.x + p.y*p.y)))
}

func testFunction() {
	const testString string = "Test TEST test"
	fmt.Println(testString)
}

func testMidValue(array []float32) float32 {
	if len(array) == 0 {
		return 0.0
	}

	var summValue float32 = 0.0

	for _, value := range array {
		summValue += value
	}
	result := summValue / float32(len(array))
	return result
}

func testOther() {
	helloString := "Hello World"
	fmt.Println(helloString)

	testFunction()

	{
		i := 0
		for i < 10 {
			fmt.Println("Value =", i)
			i += 1
		}
	}

	{
		var array [5]int
		array[2] = 10
		array[4] = 7
		fmt.Println(array)

		for _, value := range array {
			fmt.Println(value)
		}
	}

	{
		sliceSrc := make([]float32, 10)
		sliceSrc[3] = 10
		sliceSrc[4] = 20
		sliceSrc[5] = 30
		slice := sliceSrc[4:6]
		fmt.Println(sliceSrc)
		fmt.Println(slice)
	}

	{
		slice1 := []int{1, 2, 3}
		slice2 := append(slice1, 4, 5)
		fmt.Println(slice1, slice2)
	}

	{
		slice1 := []int{1, 2, 3}
		slice2 := make([]int, 2)
		copy(slice2, slice1)
		fmt.Println(slice1, slice2)
	}

	{
		testMap := make(map[string]int)
		testMap["testKey1"] = 100
		testMap["testKey2"] = 200
		fmt.Println(testMap)

		if value, ok := testMap["testKey"]; ok {
			fmt.Println(value)
		} else {
			fmt.Println("No test value")
		}
	}

	{
		testArray := []float32{10, 12, 14, 16, 12, 45}
		midValue := testMidValue(testArray)
		fmt.Println("Middle value =", midValue)
	}

	{
		pointPtr := new(Point)
		fmt.Println(pointPtr)

		var point1 Point
		fmt.Println(point1)

		point2 := Point{x: 1, y: 2}
		length := point2.length()
		fmt.Println(point2)
		fmt.Println(length)
	}
}

func testChannels() {
	chan1 := make(chan int)

	var wait sync.WaitGroup

	wait.Add(1)
	go func(c chan int) {
		chan1 <- 1
		chan1 <- 2
		chan1 <- 3
		close(chan1)

		wait.Done()
	}(chan1)

	wait.Add(1)
	go func(c chan int) {
		/*for i := range c {
			fmt.Println("Value:", i)
		}*/
		for {
			i, exists := <-c
			if exists == true {
				fmt.Println("Value:", i, exists)
			} else {
				fmt.Println("Channel closed")
				break
			}
		}
		i, exists := <-c
		fmt.Println("Value:", i, exists)

		wait.Done()
	}(chan1)

	wait.Wait()
}

func testThreads() {
	var wait sync.WaitGroup

	for i := 0; i < 128; i++ {
		wait.Add(1)
		go func() {
			var testVal float64
			for j := 0; j < 10000000; j++ {
				testVal += 4
				testVal /= 2
				testVal += (float64)(rand.Int31() % int32(3))
			}
			fmt.Println("Value:", testVal)

			wait.Done()
		}()
	}
	wait.Wait()
}

func main() {
	testThreads()
}
