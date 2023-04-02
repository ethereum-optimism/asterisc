package main

import (
	"fmt"
	"os"
)

func main() {
	_, _ = os.Stdout.Write([]byte("hello world!\n"))
	fmt.Println("printing!")
	//x := fmt.Sprintf("todo")
	fmt.Printf("formatting! %x", 123)
	//fmt.Println("completing", x)
	os.Exit(0)
}
