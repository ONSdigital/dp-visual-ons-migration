package zebedee

import (
	"time"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
	"regexp"
	"strings"
)

const (
	iframeTagRegex         = "\\[iframe url[ ]*=[ ]*.+]"
	openIFrameTagRegex     = "\\[iframe url[ ]*=[ ]*"
	closeIFrameTag         = "]"
	onsInteractiveOpenTag  = "<ons-interactive url="
	onsInteractiveCloseTag = " full-width=\"true\"/>"
	pageType               = "article"
	footnotesRegexPattern  = "\\[footnote].+\\/footnote\\]"
	footnotesOpenTag       = "[footnote]"
	footnotesCloseTag      = "[/footnote]"
	onsFootnoteIndexFMT    = "^%d^"
	onsFootnotesTitle      = "\n\n###Footnoes:"
	onsFootnoteFMT         = "\n%d. %s"
)

var (
	footnotesRegex = regexp.MustCompile(footnotesRegexPattern)
)

func CreateArticle(details *migration.Article, visualItem *gofeed.Item) *Article {
	t, err := time.Parse("02.01.06", details.PublishDate)
	if err != nil {
		panic(err)
	}

	desc := Description{
		Title:       details.Title,
		Keywords:    details.Keywords,
		ReleaseDate: "2018-01-22T00:00:00.000Z",
		Edition:     fmt.Sprintf("%s %d", t.Month().String(), t.Year()),
	}

	encoded := visualItem.Extensions["content"]["encoded"]

	section := &MarkdownSection{
		Title:    visualItem.Title,
		Markdown: encoded[0].Value,
	}
	return &Article{
		PDFTable:                  []interface{}{},
		Description:               desc,
		IsPrototypeArticle:        true,
		Sections:                  []*MarkdownSection{section},
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
		Type:                      pageType,
		Topics:                    []interface{}{},
	}
}

type Article struct {
	PDFTable                  []interface{}      `json:"pdfTable"`
	IsPrototypeArticle        bool               `json:"isPrototypeArticle"`
	Sections                  []*MarkdownSection `json:"sections"`
	Accordion                 []interface{}      `json:"accordion"`
	RelatedData               []interface{}      `json:"relatedData"`
	RelatedDocs               []interface{}      `json:"relatedDocuments"`
	Charts                    []interface{}      `json:"charts"`
	Tables                    []interface{}      `json:"tables"`
	images                    []interface{}      `json:"images"`
	Equations                 []interface{}      `json:"equations"`
	Links                     []interface{}      `json:"links"`
	RelatedMethodology        []interface{}      `json:"relatedMethodology"`
	RelatedMethodologyArticle []interface{}      `json:"relatedMethodologyArticle"`
	Versions                  []interface{}      `json:"versions"`
	Type                      string             `json:"type"`
	URI                       string             `json:"uri"`
	Description               Description        `json:"description"`
	Topics                    []interface{}      `json:"topics"`
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

func (a *Article) ConvertToONSFormat() {
	for _, section := range a.Sections {
		section.ConvertToONSFormat()
	}
}

func (s *MarkdownSection) ConvertToONSFormat() *MarkdownSection {
	s.fixInteractiveLinks()
	s.fixFootnotes()
	return s
}

func (s *MarkdownSection) fixInteractiveLinks() {
	wpIframeRE := regexp.MustCompile(iframeTagRegex)
	wpIframeOpenRE := regexp.MustCompile(openIFrameTagRegex)

	wpIframeLinks := wpIframeRE.FindAllString(s.Markdown, -1)

	for _, originalLink := range wpIframeLinks {
		updated := wpIframeOpenRE.ReplaceAllString(originalLink, onsInteractiveOpenTag)
		updated = strings.Replace(updated, closeIFrameTag, onsInteractiveCloseTag, 1)

		s.Markdown = strings.Replace(s.Markdown, originalLink, updated, 1)
	}
}

func (s *MarkdownSection) fixFootnotes() {
	fixedFootnotes := make([]string, 0)
	wpFootnotes := footnotesRegex.FindAllString(s.Markdown, -1)

	for i, wpFootnote := range wpFootnotes {
		// extract the footnote content from the tags
		florenceFootnote := strings.Replace(wpFootnote, footnotesOpenTag, "", -1)
		florenceFootnote = strings.Replace(florenceFootnote, footnotesCloseTag, "", -1)

		// replace the tags with the index of the footnote
		s.Markdown = strings.Replace(s.Markdown, wpFootnote, fmt.Sprintf(onsFootnoteIndexFMT, i+1), -1)
		fixedFootnotes = append(fixedFootnotes, florenceFootnote)
	}

	onsFootnotes := onsFootnotesTitle
	for i, fn := range fixedFootnotes {
		onsFootnotes += fmt.Sprintf(onsFootnoteFMT, i+1, fn)
	}

	s.Markdown += onsFootnotes
}
