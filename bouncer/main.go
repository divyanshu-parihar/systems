package main

import (
	"fmt"
	"log/slog"

	cmd "github.com/divyanshu-parihar/systems/bouncer/cmd"
)

func main() {
	fmt.Println("Hello, World")
	err := cmd.Run()
	if err != nil {
		slog.Error("Error running bouncer")
		return
	}
	fmt.Println("Started the bouncer")
}
