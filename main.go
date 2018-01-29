package main

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/zebedee"
	"os"
	"github.com/pkg/errors"
	"github.com/ONSdigital/dp-visual-ons-migration/config"
	"flag"
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
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

	migrationArticles, err := migration.ParseMigrationFile(cfg.MigrationFile)
	if err != nil {
		exit(err)
	}

	visualMapping, err := migration.ParseRSSFeed(cfg.VisualRSSFile)
	if err != nil {
		exit(err)
	}

	for _, article := range migrationArticles {
		if _, ok := visualMapping[article.VisualURL]; !ok {
			err := errors.New("visual entry was not found in rss mapping")
			log.Error(err, log.Data{"visualURL": article.VisualURL})
			exit(err)
		}

		visualItem := visualMapping[article.VisualURL]

		col, err := zebedee.CreateCollection(article.Title)
		if err != nil {
			exit(err)
		}

		a := zebedee.CreateArticle(article, visualItem)
		a.ConvertToONSFormat()

		if err := col.AddArticle(a, article); err != nil {
			exit(err)
		}
	}
}

func exit(err error) {
	migrationErr, ok := err.(migration.Error)
	if ok {
		log.Error(migrationErr, migrationErr.Params)
	} else {
		log.Error(err, nil)
	}
	os.Exit(1)
}
