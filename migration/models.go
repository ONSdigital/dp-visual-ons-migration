package migration

import (
	"github.com/ONSdigital/go-ns/log"
	"regexp"
	"strings"
)

type Error struct {
	Message     string
	OriginalErr error
	Params      log.Data
}

type Article struct {
	PublishDate  string   `json:"publishDate"`
	PostTitle    string   `json:"postTitle"`
	Title        string   `json:"title"`
	TaxonomyURI  string   `json:"taxonomyURI"`
	RelatedLinks []string `json:"relatedLinks"`
	Keywords     []string `json:"keywords"`
	VisualURL    string   `json:"visualURL"`
}

func (e Error) Error() string {
	return e.Message + ": " + e.OriginalErr.Error()
}

func (m *Article) GetTaxonomyURI() string {
	return m.TaxonomyURI + "/" + strings.TrimSpace(strings.ToLower(m.GetCollectionName()))
}

func (m *Article) GetCollectionName() string {
	r, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(r.ReplaceAllString(m.Title, ""))
}
