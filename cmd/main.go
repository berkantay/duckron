package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/berkantay/duckron"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	configReader := duckron.NewConfigReader()
	config, err := configReader.Read()
	if err != nil {
		log.Fatal(err)
	}

	duckron, err := duckron.NewDuckron(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Duckron started with", config)

	go func() {
		err := duckron.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-sigChan
	fmt.Println("Received interrupt signal, shutting down...")
}
