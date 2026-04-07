package main

import (
	"log"

	"github.com/djengua/mqtt-ingestor/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
