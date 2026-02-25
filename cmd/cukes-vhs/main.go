package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("cukes-vhs version", version)
		return
	}

	fmt.Println("cukes-vhs - VHS recording tool for Cucumber scenarios")
}
