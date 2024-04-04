package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lucasmbaia/power-actions/config"
	"github.com/lucasmbaia/power-actions/core"
)

func init() {
	config.LoadSingletons()
}

func main() {
	var (
		err      error
		diffPath string
	)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go <diff_path>")
		os.Exit(1)
	}
	diffPath = os.Args[1]

	if err = core.Run(diffPath); err != nil {
		log.Fatalf("Error when running script: %s", err.Error())
	}
}
