package main

import (
	"github.com/carousell/ct-go/pkg/logger"
	"github.com/ct-logic-api-document/cmd"
)

func main() {
	err := cmd.Execute()
	log := logger.MustNamed("app")
	if err != nil {
		log.Errorf("application execute err: %s", err)
	}
}
