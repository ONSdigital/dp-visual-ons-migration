package mapping

import (
	"os"
	"encoding/csv"
	"io"
	"github.com/ONSdigital/go-ns/log"
	"encoding/json"
	"strings"
	"regexp"
)

type MigrationDetails struct {
	PublishDate  string   `json:"publishDate"`
	PostTitle    string   `json:"postTitle"`
	Title        string   `json:"title"`
	TaxonomyURI  string   `json:"taxonomyURI"`
	RelatedLinks []string `json:"relatedLinks"`
	Keywords     []string `json:"keywords"`
	VisualURL    string   `json:"visualURL"`
}

func ParseMapping(filename string) (*MigrationDetails, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
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
			log.Error(err, nil)
		}

		if isHeader {
			log.Info("skipping header row", nil)
			isHeader = false
			continue
		}

		mapping = append(mapping, row)
	}

	line := mapping[0]
	m := &MigrationDetails{
		PublishDate:  strings.TrimSpace(line[0]),
		PostTitle:    strings.TrimSpace(line[1]),
		Title:        strings.TrimSpace(line[2]),
		TaxonomyURI:  strings.TrimSpace(line[3]),
		RelatedLinks: toSlice(line[4], ","),
		Keywords:     toSlice(line[5], ";"),
		VisualURL:    strings.TrimSpace(line[6]),
	}

	b, err := json.MarshalIndent(m, "", "	")
	if err != nil {
		return nil, err
	}
	log.Debug("record", log.Data{"->": string(b)})

	return m, nil
}

func (m *MigrationDetails) GetTaxonomyURI() string {
	return m.TaxonomyURI + "/" + strings.TrimSpace(strings.ToLower(m.GetCollectionName()))
}

func (m *MigrationDetails) GetCollectionName() string {
	r, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(r.ReplaceAllString(m.Title, ""))
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
