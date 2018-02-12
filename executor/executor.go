package executor

import (
	"os"
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
	"encoding/csv"
	"strconv"
	"github.com/mmcdole/gofeed"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-visual-ons-migration/zebedee"
)

const (
	migrationErrorsFile = "migration-errors.csv"
	entryNotFound       = "visual entry was not found in rss mapping"
	conversionErr       = "error while attempting to convert visual post to collection article"
	resultsFilename     = "migration-results.csv"
)

var (
	migrationErrorsFileHeader = []string{"MAPPING_INDEX", "ERROR", "VISUAL_URL", "ONS_URL", "COLLECTION"}
	resultsFileHeader         = []string{"MAPPING_INDEX", "COLLECTION_NAME", "STATUS", "VISUAL_URL", "ONS_URL"}
)

type Executor struct {
	plan            *migration.Plan
	startIndex      int
	currentRowIndex int
	errorsCount     int
	errFile         *os.File
	resultsFile     *os.File
	errWriter       *csv.Writer
	resultsWriter   *csv.Writer
}

func newFile(name string) (*os.File, error) {
	var f *os.File
	if _, err := os.Stat(name); os.IsNotExist(err) {
		f, _ = os.Create(name)
	} else {
		os.Remove(name)
		f, err = os.Create(name)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

func New(plan *migration.Plan, startIndex int) (*Executor, error) {
	errFile, _ := newFile(migrationErrorsFile)
	resultsFile, _ := newFile(resultsFilename)

	errWriter := csv.NewWriter(errFile)
	errWriter.Write(migrationErrorsFileHeader)

	resultsWriter := csv.NewWriter(resultsFile)
	resultsWriter.Write(resultsFileHeader)

	return &Executor{plan: plan,
		errorsCount: 0,
		currentRowIndex: startIndex,
		errWriter: errWriter,
		errFile: errFile,
		resultsFile: resultsFile,
		resultsWriter: resultsWriter,
	}, nil
}

func (e *Executor) Migrate(start int, batchSize int) {
	end := start + batchSize

	if end > len(e.plan.Mapping.ArticleURLsOrdered) {
		log.Debug("batch size exceeds input total input length, reducing batch size", log.Data{
			"original": batchSize,
			"new":      len(e.plan.Mapping.ArticleURLsOrdered),
		})
		batchSize = len(e.plan.Mapping.ArticleURLsOrdered)
	}

	// use the list to maintain the order in which the entries appear in the file
	log.Info("processing batch", log.Data{"start": start, "end": end})

	batch := e.plan.Mapping.ArticleURLsOrdered[start:end]

	for _, migrationURL := range batch {
		article, _ := e.plan.Mapping.ToMigrate[migrationURL]

		if err := article.Valid(); err != nil {
			e.logError(err, article, "")
			continue
		}

		var visualItem *gofeed.Item
		var ok bool

		if visualItem, ok = e.plan.VisualExport.Posts[article.VisualURL]; !ok {
			err := migration.Error{Message: entryNotFound, OriginalErr: nil, Params: log.Data{"visualURL": article.VisualURL}}
			e.logError(err, article, "")
			continue
		}

		collectionName := zebedee.ToCollectionName(e.currentRowIndex, article.PostTitle)

		col, err := zebedee.CreateCollection(collectionName)
		if err != nil {
			e.logError(err, article, collectionName)
			continue
		}

		a := zebedee.CreateArticle(article, visualItem)
		if err := a.ConvertToONSFormat(e.plan); err != nil {
			e.logError(migration.Error{Message: conversionErr, OriginalErr: err, Params: log.Data{"title": visualItem.Title}}, article, collectionName)
			continue
		}

		if err := col.AddArticle(a, article); err != nil {
			e.logError(err, article, collectionName)
			continue
		}
		e.logMigrationOutcome(nil, article, collectionName)
	}
}

func (e *Executor) logMigrationOutcome(err error, article *migration.Article, collectionName string) {
	status := "success"
	if err != nil {
		status = "error"
	}
	e.resultsWriter.Write([]string{strconv.Itoa(e.currentRowIndex), collectionName, status, article.VisualURL, article.TaxonomyURI})
	e.currentRowIndex++
}

func (e *Executor) logError(err error, post *migration.Article, collectionName string) {
	log.ErrorC("error while processing mapping entry", err, log.Data{"rowIndex": e.currentRowIndex})
	e.errWriter.Write([]string{strconv.Itoa(e.currentRowIndex), err.Error(), post.VisualURL, post.TaxonomyURI, collectionName})
	e.errorsCount++

	e.logMigrationOutcome(err, post, collectionName)
}

func (e *Executor) Close() {
	log.Info("closing executor resources", nil)
	e.errWriter.Flush()
	e.resultsWriter.Flush()
	e.errFile.Close()
	e.resultsFile.Close()
}
