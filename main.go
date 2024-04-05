package main

import (
	"log"

	"github.com/lucasmbaia/power-actions/config"
	"github.com/lucasmbaia/power-actions/core"
)

func init() {
	config.LoadSingletons()
}

func main() {
	if err := core.Run(); err != nil {
		log.Fatalf("Error when running script: %s", err.Error())
	}
}
