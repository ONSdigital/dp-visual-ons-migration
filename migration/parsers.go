package migration

import (
	"os"
	"io"
	"encoding/csv"
	"github.com/ONSdigital/go-ns/log"
	"strings"
	"github.com/mmcdole/gofeed"
	"github.com/ONSdigital/dp-visual-ons-migration/config"
	"fmt"
	"github.com/ONSdigital/dp-visual-ons-migration/util"
)

const (
	staticONSHost    = "https://static.ons.gov.uk"
	ONSSite          = "https://www.ons.gov.uk"
	wpAttachmentPath = "/wp-content/uploads/"
	staticONSPath    = "/visual/"
	postType         = "post"
	attachmentType   = "attachment"
)

func LoadPlan(cfg *config.Model) (*Plan, error) {

	migrationMapping, err := parseMappingFile(cfg.MappingFile)
	if err != nil {
		return nil, err
	}

	visualExport, err := parseVisualExport(cfg.VisualExportFile, migrationMapping)
	if err != nil {
		return nil, err
	}

	return &Plan{
		Mapping:             migrationMapping,
		VisualExport:        visualExport,
		NationalArchivesURL: cfg.NationalArchivesURL,
	}, nil
}

// Parse the mapping file.
func parseMappingFile(filename string) (*Mapping, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, Error{"error while attempting to open migration file", err, log.Data{"filename": filename}}
	}

	defer f.Close()

	rows := make([][]string, 0)
	mapping := &Mapping{ToMigrate: make([]*Article, 0)}
	isHeader := true
	reader := csv.NewReader(f)

	for {
		row, err := reader.Read()

		if err == io.EOF {
			log.Info("end of csv reached", nil)
			break
		}

		if err != nil {
			return nil, Error{"error while reading migration file", err, nil}
		}

		if isHeader {
			isHeader = false
			continue
		}

		rows = append(rows, row)
	}

	for _, line := range rows {

		a := &Article{
			PostTitle:    strings.TrimSpace(line[0]),
			TaxonomyURI:  strings.TrimSpace(line[1]),
			RelatedLinks: []string{},
			Keywords:     toSlice(line[3], ";"),
			VisualURL:    strings.TrimSpace(line[4]),
		}

		mapping.ToMigrate = append(mapping.ToMigrate, a)
	}

	return mapping, nil
}

// parse the visual ons rss file into the visual export structure
func parseVisualExport(filename string, m *Mapping) (*VisualExport, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	vm := newVisualExport()

	//log.Info("attempting to parse RSS export file", nil)
	fp := gofeed.NewParser()
	visualFeed, err := fp.Parse(file)
	if err != nil {
		return nil, Error{"failed to parse visual RSS feed", err, nil}
	}

	//log.Info("mapping visual posts by post url", nil)
	for _, item := range visualFeed.Items {
		t := item.Extensions["wp"]["post_type"][0]

		if attachmentType == t.Value {
			vm.addAttachment(item)
		} else if postType == t.Value {
			vm.addPost(item)

			if a, ok := m.GetArticleByURL(item.Link); ok {
				//log.Info("adding post to migration mapping", log.Data{"visualURL": item.Link})
				vm.addPost(item)
				a.TaxonomyURI = fmt.Sprintf("%s/articles/%s/%s", a.TaxonomyURI, util.SanitisedFilename(item.Title), item.PublishedParsed.Format("2006-01-02"))
			}
		}
	}
	//log.Info("mapping generated successfully", nil)
	return vm, nil
}

func toSlice(line string, delimiter string) []string {
	if len(line) == 0 {
		return nil
	}

	if !strings.Contains(line, delimiter) {
		return nil
	}

	items := strings.SplitN(line, delimiter, -1)
	for i, val := range items {
		items[i] = strings.TrimSpace(val)
	}
	return items
}
