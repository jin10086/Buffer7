package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: buffer7 <command> [args...]")
		os.Exit(1)
	}
	fmt.Printf("Buffer7 is wrapping: %v\n", os.Args[1:])
}
