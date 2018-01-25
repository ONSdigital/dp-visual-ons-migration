package main

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/mapping"
	"github.com/ONSdigital/dp-visual-ons-migration/zebedee"
	"os"
	"github.com/pkg/errors"
	"github.com/ONSdigital/dp-visual-ons-migration/config"
	"github.com/ONSdigital/dp-visual-ons-migration/visual"
	"flag"
)

func main() {
	log.HumanReadable = true
	log.Info("dp-visual-migration", nil)

	cfgFile := flag.String("cfg", "config.yml", "the config to use when running the migration")
	flag.Parse()

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		exit(errors.Wrap(err, "failed loading config"))
	}

	log.Info("configuring collections root directory", log.Data{"dir": cfg.CollectionsDir})
	zebedee.CollectionsRoot = cfg.CollectionsDir

	migrationMapping, err := mapping.ParseMigrationFile(cfg.MigrationFile)
	if err != nil {
		exit(err)
	}

	visualMapping, err := visual.ParseRSSFeed(cfg.VisualRSSFile)
	if err != nil {
		exit(err)
	}

	for _, migrationDetail := range migrationMapping {
		if _, ok := visualMapping[migrationDetail.VisualURL]; !ok {
			err := errors.New("visual entry was not found in rss mapping")
			log.Error(err, log.Data{"visualURL": migrationDetail.VisualURL})
			exit(err)
		}

		visualItem := visualMapping[migrationDetail.VisualURL]

		col, err := zebedee.CreateCollection(migrationDetail.Title)
		if err != nil {
			exit(err)
		}

		a := zebedee.CreateArticle(migrationDetail, visualItem)
		if err := col.AddArticle(a, migrationDetail); err != nil {
			exit(err)
		}
	}
}

func exit(err error) {
	log.Error(err, nil)
	os.Exit(1)
}
