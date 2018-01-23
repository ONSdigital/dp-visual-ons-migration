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
	flag.Parse()

	if len(*collectionsDir) == 0 {
		panic(errors.New("no collections directory path was provided"))
	}

	zebedee.CollectionsRoot = *collectionsDir

	m, err := mapping.ParseMapping("example-mapping.csv")
	if err != nil {
		panic(errors.New("failed to parse migration mapping file"))
		os.Exit(1)
	}

	c, err := zebedee.CreateCollection(m.GetCollectionName())
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	a := zebedee.CreateArticle(m)
	c.AddArticle(a, m)
}
