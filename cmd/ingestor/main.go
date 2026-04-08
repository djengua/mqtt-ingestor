package main

import (
	"fmt"
	"log"

	"github.com/djengua/mqtt-ingestor/internal/app"
)

func main() {
	fmt.Println("MAIN START")
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
