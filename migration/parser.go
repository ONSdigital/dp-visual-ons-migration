package migration

import (
	"os"
	"io"
	"encoding/csv"
	"github.com/ONSdigital/go-ns/log"
	"strings"
)

func ParseMigrationFile(filename string) ([]*Article, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, Error{"error while attempting to open migration file", err, log.Data{"filename": filename}}
	}

	defer f.Close()

	mapping := make([][]string, 0)
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

		mapping = append(mapping, row)
	}

	migrationPlan := make([]*Article, 0)
	for _, line := range mapping {
		migrationPlan = append(migrationPlan, &Article{
			PublishDate:  strings.TrimSpace(line[0]),
			PostTitle:    strings.TrimSpace(line[1]),
			Title:        strings.TrimSpace(line[2]),
			TaxonomyURI:  strings.TrimSpace(line[3]),
			RelatedLinks: toSlice(line[4], ","),
			Keywords:     toSlice(line[5], ";"),
			VisualURL:    strings.TrimSpace(line[6]),
		})
	}

	return migrationPlan, nil
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
