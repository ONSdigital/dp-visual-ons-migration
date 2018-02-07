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
	"encoding/csv"
	"strconv"
)

var errWriter *csv.Writer

const (
	migrationErrorsFile = "migration-errors.csv"
	entryNotFound       = "visual entry was not found in rss mapping"
	conversionErr       = "error while attempting to convert visual post to collection article"
)

type Executor struct {
	plan            *migration.Plan
	startIndex      int
	endIndex        int
	currentRowIndex int
	errorsCount     int
}

func main() {
	log.HumanReadable = true
	log.Info("dp-visual-migration", nil)

	cfgFile := flag.String("cfg", "config.yml", "the config to use when running the migration")
	flag.Parse()

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		exit(errors.Wrap(err, "failed loading config"))
	}

	var cleanUp func()
	errWriter, cleanUp = getErrRecorder()
	defer cleanUp()

	log.Info("configuring collections root directory", log.Data{"dir": cfg.CollectionsDir})
	zebedee.CollectionsRoot = cfg.CollectionsDir

	plan, err := migration.LoadPlan(cfg.MappingFile, cfg.VisualExportFile)
	if err != nil {
		exit(err)
	}

	e := &Executor{plan: plan, startIndex: 0, endIndex: 9, errorsCount: 0, currentRowIndex: 0}
	e.migrateArticles()

	if e.errorsCount != 0 {
		log.Error(errors.New("errors were encountered while running plan, see migration-errors.csv for me details"), log.Data{
			"errorCount": e.errorsCount,
		})
	} else {
		log.Info("no errors were encountered", nil)
	}
}

func (e *Executor) migrateArticles() {
	for _, article := range e.plan.Mapping.PostsToMigrate {

		if err := article.Valid(); err != nil {
			e.trackError(err, article, "")
			continue
		}

		var visualItem *gofeed.Item
		var ok bool

		if visualItem, ok = e.plan.VisualExport.Posts[article.VisualURL]; !ok {
			err := migration.Error{Message: entryNotFound, OriginalErr: nil, Params: log.Data{"visualURL": article.VisualURL}}
			e.trackError(err, article, "")
			continue
		}

		col, err := zebedee.CreateCollection(article.PostTitle)
		if err != nil {
			e.trackError(err, article, "")
			continue
		}

		a := zebedee.CreateArticle(article, visualItem)
		if err := a.ConvertToONSFormat(e.plan); err != nil {
			e.trackError(migration.Error{Message: conversionErr, OriginalErr: err, Params: log.Data{"title": visualItem.Title}}, article, col.Name)
			continue
		}

		if err := col.AddArticle(a, article); err != nil {
			e.trackError(err, article, col.Name)
			continue
		}
		e.currentRowIndex += 1
	}
}

func (e *Executor) trackError(err error, post *migration.Article, collectionName string) {
	errWriter.Write([]string{strconv.Itoa(e.currentRowIndex), err.Error(), post.VisualURL, post.TaxonomyURI, collectionName})
	e.currentRowIndex++
	e.errorsCount++
}

func getErrRecorder() (*csv.Writer, func()) {
	var f *os.File
	if _, err := os.Stat(migrationErrorsFile); os.IsNotExist(err) {
		f, _ = os.Create(migrationErrorsFile)
	} else {
		os.Remove(migrationErrorsFile)
		f, _ = os.Create(migrationErrorsFile)
	}

	w := csv.NewWriter(f)
	w.Write([]string{"INDEX", "ERROR", "VISUAL_URL", "ONS_URL", "COLLECTION"})
	return w, func() {
		errWriter.Flush()
		f.Close()
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
