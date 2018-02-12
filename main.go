package main

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/zebedee"
	"os"
	"github.com/pkg/errors"
	"github.com/ONSdigital/dp-visual-ons-migration/config"
	"flag"
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
	"github.com/ONSdigital/dp-visual-ons-migration/executor"
)

func main() {
	log.HumanReadable = true
	log.Info("dp-visual-migration", nil)

	cfgFile := flag.String("cfg", "local.yml", "the config to use when running the migration")
	startIndex := flag.Int("start", 0, "")
	flag.Parse()

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		exit(errors.Wrap(err, "failed loading config"))
	}

	log.Info("configuring collections root directory", log.Data{"dir": cfg.CollectionsDir})
	zebedee.CollectionsRoot = cfg.CollectionsDir

	plan, err := migration.LoadPlan(cfg, *startIndex)
	if err != nil {
		exit(err)
	}

	e, err := executor.New(plan, *startIndex)
	if err != nil {
		exit(err)
	}
	defer e.Close()

	e.Migrate()
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
