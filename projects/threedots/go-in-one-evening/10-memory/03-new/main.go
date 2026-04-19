package main

import "fmt"

var counter int

func AllocateBuffer() *string {
	if counter >= 3 {
		return nil
	}

	counter++
	return new(string)
}

func main() {
	var buffers []*string

	for {
		b := AllocateBuffer()
		if b == nil {
			break
		}

		buffers = append(buffers, b)
	}

	fmt.Println("Allocated", len(buffers), "buffers")
}
