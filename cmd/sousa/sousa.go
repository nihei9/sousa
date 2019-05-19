package main

import (
	"os"

	"github.com/nihei9/sousa/pkg/sousa/cmd"
)

func main() {
	os.Exit(run())
}

func run() int {
	cmd := cmd.NewCmd()
	cmd.SetOutput(os.Stdout)
	err := cmd.Execute()
	if err != nil {
		cmd.SetOutput(os.Stderr)
		cmd.Println(err)
		return 1
	}

	return 0
}
