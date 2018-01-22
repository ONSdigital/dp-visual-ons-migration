package main

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/collections"
	"os"
	"flag"
)

func main() {
	log.HumanReadable = true
	log.Info("dp-visual-migration", nil)

	collectionsDir := flag.String("collections", "", "zebedee collections dir")
	flag.Parse()

	c := collections.New("dave-test")
	if err := c.Create(*collectionsDir); err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}
