package main

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/zebedee"
	"os"
	"github.com/pkg/errors"
	"github.com/ONSdigital/dp-visual-ons-migration/config"
	"flag"
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
	"github.com/mmcdole/gofeed"
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

	plan, err := migration.LoadPlan(cfg.MappingFile, cfg.VisualExportFile)
	if err != nil {
		exit(err)
	}

	for _, post := range plan.Mapping.PostsToMigrate {

		var visualItem *gofeed.Item
		var ok bool

		if visualItem, ok = plan.VisualExport.Posts[post.VisualURL]; !ok {
			err := errors.New("visual entry was not found in rss mapping")
			log.Error(err, log.Data{"visualURL": post.VisualURL})
			exit(err)
		}

		col, err := zebedee.CreateCollection(post.Title)
		if err != nil {
			exit(err)
		}

		a := zebedee.CreateArticle(post, visualItem)
		if err := a.ConvertToONSFormat(plan); err != nil {
			log.ErrorC("error while attempting to convert visual post to collection post", err, log.Data{"title": visualItem.Title})
			exit(err)
		}

		if err := col.AddArticle(a, post); err != nil {
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
