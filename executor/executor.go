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
	entryNotFound = "visual url entry was not found in this version of the wordpress export mapping"
	conversionErr = "error while attempting to convert visual post to collection article"
)

var (
	resultsFileHeader = []string{"MAPPING_ROW_INDEX", "COLLECTION_NAME", "STATUS", "VISUAL_URL", "ONS_URL", "ERROR_DETAILS"}
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

func New(plan *migration.Plan, startIndex int, resultsPath string) (*Executor, error) {
	resultsFile, _ := newFile(resultsPath)
	resultsWriter := csv.NewWriter(resultsFile)
	resultsWriter.Write(resultsFileHeader)

	return &Executor{plan: plan,
		errorsCount: 0,
		currentRowIndex: startIndex,
		resultsFile: resultsFile,
		resultsWriter: resultsWriter,
	}, nil
}

func (e *Executor) Migrate(start int, batchSize int) {
	end := start + batchSize

	if end > len(e.plan.Mapping.ToMigrate) {
		log.Debug("batch size exceeds input total input length, reducing batch size", log.Data{
			"original": batchSize,
			"new":      len(e.plan.Mapping.ToMigrate),
		})
		batchSize = len(e.plan.Mapping.ToMigrate)
	}

	// use the list to maintain the order in which the entries appear in the file
	log.Info("processing batch", log.Data{"start": start, "end": end})

	batch := e.plan.Mapping.ToMigrate[start:end]

	for _, article := range batch {

		if err := article.Valid(); err != nil {
			e.logMigrationOutcome(err, article.VisualURL, "", "")
			continue
		}

		var visualItem *gofeed.Item
		var ok bool

		if visualItem, ok = e.plan.VisualExport.Posts[article.VisualURL]; !ok {
			err := migration.Error{Message: entryNotFound, OriginalErr: nil, Params: log.Data{"visualURL": article.VisualURL}}
			e.logMigrationOutcome(err, article.VisualURL, "", "")
			continue
		}

		// offset the index to align with the input mapping - the header row is ignored and array is indexed from 0
		collectionName := zebedee.ToCollectionName(e.currentRowIndex+2, article.PostTitle)

		col, err := zebedee.CreateCollection(collectionName)
		if err != nil {
			e.logMigrationOutcome(err, article.VisualURL, "", collectionName)
			continue
		}

		a := zebedee.CreateArticle(article, visualItem)
		if err := a.ConvertToONSFormat(e.plan); err != nil {
			err := migration.Error{Message: conversionErr, OriginalErr: err, Params: log.Data{"title": visualItem.Title}}
			e.logMigrationOutcome(err, article.VisualURL, a.URI, collectionName)
			continue
		}

/*		if uri := e.plan.VisualExport.GetThumbnailURL(a.ImageURI); uri != nil {
			e.plan.GetMigratedURL(uri.String())
			imgURI, err := e.plan.GetMigratedURL(uri.String())
			if err != nil {
				e.logMigrationOutcome(err, article.VisualURL, a.URI, collectionName)
				continue
			}
			a.ImageURI = imgURI
		}*/

		if err := col.AddArticle(a, article); err != nil {
			e.logMigrationOutcome(err, article.VisualURL, a.URI, collectionName)
			continue
		}
		e.logMigrationOutcome(nil, article.VisualURL, a.URI, collectionName)
	}
}

func (e *Executor) logMigrationOutcome(err error, visualURL string, onsURL string, collectionName string) {
	status := "SUCCESS"
	errMsg := "N/A"
	if err != nil {
		log.ErrorC("error while processing mapping entry", err, log.Data{"rowIndex": e.currentRowIndex})
		errMsg = err.Error()
		status = "ERROR"
	}

	// take into account the header row of the csv + array being indexed from 0
	index := e.currentRowIndex + 2
	e.resultsWriter.Write([]string{strconv.Itoa(index), collectionName, status, visualURL, onsURL, errMsg})
	e.currentRowIndex++
}

func (e *Executor) Close() {
	log.Debug("closing executor resources", nil)
	e.resultsWriter.Flush()
	e.resultsFile.Close()
}
