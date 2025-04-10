package main

import (
	"fmt"
	"os"

	"github.com/github/gh-combine/internal/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
