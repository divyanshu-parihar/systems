package main

import (
	"fmt"

	cmd "github.com/divyanshu-parihar/systems/bouncer/cmd"
)

func main() {
	fmt.Println("Hello, World")
	err := cmd.Run()
	if err != nil {
		fmt.Errorf("Error running bouncer")
		return
	}
	fmt.Println("Started the bouncer")
}
