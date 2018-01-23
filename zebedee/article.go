package zebedee

import (
	"github.com/ONSdigital/dp-visual-ons-migration/mapping"
)

func CreateArticle(details *mapping.MigrationDetails) *Article {
	desc := Description{
		Title:       details.Title,
		Keywords:    details.Keywords,
		ReleaseDate: "2018-01-22T00:00:00.000Z",
	}

	//encoded := item.Extensions["content"]["encoded"]

	section := MarkdownSection{
		Title:    "Section 1",
		Markdown: "",
	}
	return &Article{
		PDFTable:                  []interface{}{},
		Description:               desc,
		IsPrototypeArticle:        true,
		Sections:                  []MarkdownSection{section},
		Accordion:                 []interface{}{},
		RelatedData:               []interface{}{},
		RelatedDocs:               []interface{}{},
		Charts:                    []interface{}{},
		Tables:                    []interface{}{},
		Equations:                 []interface{}{},
		Links:                     []interface{}{},
		RelatedMethodology:        []interface{}{},
		RelatedMethodologyArticle: []interface{}{},
		Versions:                  []interface{}{},
		URI:                       details.GetTaxonomyURI(),
		Type:                      "article",
		Topics:                    []interface{}{},
	}
}

type Article struct {
	PDFTable                  []interface{}     `json:"pdfTable"`
	IsPrototypeArticle        bool              `json:"isPrototypeArticle"`
	Sections                  []MarkdownSection `json:"sections"`
	Accordion                 []interface{}     `json:"accordion"`
	RelatedData               []interface{}     `json:"relatedData"`
	RelatedDocs               []interface{}     `json:"relatedDocuments"`
	Charts                    []interface{}     `json:"charts"`
	Tables                    []interface{}     `json:"tables"`
	images                    []interface{}     `json:"images"`
	Equations                 []interface{}     `json:"equations"`
	Links                     []interface{}     `json:"links"`
	RelatedMethodology        []interface{}     `json:"relatedMethodology"`
	RelatedMethodologyArticle []interface{}     `json:"relatedMethodologyArticle"`
	Versions                  []interface{}     `json:"versions"`
	Type                      string            `json:"type"`
	URI                       string            `json:"uri"`
	Description               Description       `json:"description"`
	Topics                    []interface{}     `json:"topics"`
}

type MarkdownSection struct {
	Title    string `json:"title"`
	Markdown string `json:"markdown"`
}

type Description struct {
	Title             string   `json:"title"`
	Keywords          []string `json:"keywords"`
	MetaDescription   string   `json:"metaDescription"`
	NationalStatistic bool     `json:"nationalStatistic"`
	LatestRelease     bool     `json:"latestRelease"`
	Contact           Contact  `json:"contact"`
	ReleaseDate       string   `json:"releaseDate"`
	NextRelease       string   `json:"nextRelease"`
	Edition           string   `json:"edition"`
	Abstraction       string   `json:"_abstract"`
	Unit              string   `json:"unit"`
	PreUnit           string   `json:"preUnit"`
	Source            string   `json:"source"`
}

type Contact struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"telephone"`
}
