package main

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/mapping"
	"flag"
	"github.com/ONSdigital/dp-visual-ons-migration/zebedee"
	"os"
	"github.com/pkg/errors"
)

func main() {
	log.HumanReadable = true
	log.Info("dp-visual-migration", nil)

	collectionsDir := flag.String("collectionsDir", "/content/collections", "zebedee zebedee dir")
	//collectionsDir := flag.String("collectionsDir", "/Users/dave/Desktop/zebedee-data/content/zebedee/collections", "zebedee zebedee dir")
	flag.Parse()

	if len(*collectionsDir) == 0 {
		exit(errors.New("no collections directory path was provided"))
	}

	zebedee.CollectionsRoot = *collectionsDir

	m, err := mapping.ParseMapping("example-mapping.csv")
	if err != nil {
		exit(err)
	}

	c, err := zebedee.CreateCollection(m.Title)
	if err != nil {
		exit(err)
	}

	a := zebedee.CreateArticle(m)
	if err := c.AddArticle(a, m); err != nil {
		exit(err)
	}
}

func exit(err error) {
	log.Error(err, nil)
	os.Exit(1)
}
