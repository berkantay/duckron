package main

import (
	"log"

	"github.com/berkantay/duckron"
)

func main() {

	configReader := duckron.NewConfigReader()
	config, err := configReader.Read()
	if err != nil {
		log.Fatal(err)
	}

	duckron, err := duckron.NewDuckron(config)
	if err != nil {
		log.Fatal(err)
	}
	duckron.RunSnapshotJob()

}
