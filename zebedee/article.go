package zebedee

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
	"strings"
	"golang.org/x/net/html"
	"unicode"
	"strconv"
	"sort"
)

const (
	pageType             = "article"
	collectionDateFormat = "2006-01-02T03:04:05.000Z"
	visualDateFormat     = "Mon, 02 Jan 2006 03:02:05 Z0700"
	hrefTag              = "href"
	articleDateFormat    = "2006-01-02"
)

func CreateArticle(details *migration.Article, visualItem *gofeed.Item) *Article {

	desc := Description{
		Title:       details.PostTitle,
		Keywords:    details.Keywords,
		ReleaseDate: visualItem.PublishedParsed.Format(collectionDateFormat),
	}

	encoded := visualItem.Extensions["content"]["encoded"]

	section := &MarkdownSection{
		Markdown: encoded[0].Value,
	}

	links := make([]*RelatedLink, 0)

	var thumbnailID string
	links, thumbnailID = ParseMetadata(visualItem)
	sort.Slice(links, func(i, j int) bool {
		return links[i].ID < links[j].ID
	})

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
		Links:                     links,
		RelatedMethodology:        []interface{}{},
		RelatedMethodologyArticle: []interface{}{},
		Versions:                  []interface{}{},
		URI:                       details.TaxonomyURI,
		Type:                      pageType,
		Topics:                    []interface{}{},
		ImageURI:                  thumbnailID,
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
	Links                     []*RelatedLink     `json:"links"`
	RelatedMethodology        []interface{}      `json:"relatedMethodology"`
	RelatedMethodologyArticle []interface{}      `json:"relatedMethodologyArticle"`
	Versions                  []interface{}      `json:"versions"`
	Type                      string             `json:"type"`
	URI                       string             `json:"uri"`
	Description               Description        `json:"description"`
	Topics                    []interface{}      `json:"topics"`
	ImageURI                  string             `json:"imageURI"`
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

type RelatedLink struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
	ID    int    `json:"-"`
}

func ParseMetadata(visualItem *gofeed.Item) ([]*RelatedLink, string) {
	metadata := visualItem.Extensions["wp"]["postmeta"]
	rawLinks := map[int]*RelatedLink{}
	thumbnailID := ""

	for _, mi := range metadata {

		metaKey := mi.Children["meta_key"][0]
		if moreInfoURLRX.MatchString(metaKey.Value) {
			index := metaKey.Value
			index = strings.Replace(index, "more_information_", "", 1)
			index = strings.Replace(index, "_url", "", 1)
			i, _ := strconv.Atoi(index)

			if l, ok := rawLinks[i]; ok {
				l.URI = mi.Children["meta_value"][0].Value
			} else {
				rawLinks[i] = &RelatedLink{ID: i, URI: mi.Children["meta_value"][0].Value}
			}
		} else if moreInfoTitleRX.MatchString(metaKey.Value) {
			index := metaKey.Value
			index = strings.Replace(index, "more_information_", "", 1)
			index = strings.Replace(index, "_link_title", "", 1)
			i, _ := strconv.Atoi(index)

			if l, ok := rawLinks[i]; ok {
				l.Title = mi.Children["meta_value"][0].Value
			} else {
				rawLinks[i] = &RelatedLink{ID: i, Title: mi.Children["meta_value"][0].Value}
			}
		} else if metaKey.Value == "_thumbnail_id" {
			thumbnailID = mi.Children["meta_value"][0].Value
		}
	}

	links := make([]*RelatedLink, 0)
	for _, l := range rawLinks {
		links = append(links, l)
	}
	return links, thumbnailID
}

func (a *Article) ConvertToONSFormat(plan *migration.Plan) error {
	for _, s := range a.Sections {
		markdown, err := ConvertHTMLToONSMarkdown(s.Markdown, plan)
		if err != nil {
			return err
		}

		s.Markdown = markdown
		s.fixInteractiveLinks()
		s.fixFootnotes()
		s.fixExplanations()
	}
	return nil
}

func (s *MarkdownSection) fixInteractiveLinks() {
	wpIframeLinks := wpIFrameRX.FindAllString(s.Markdown, -1)

	for _, originalLink := range wpIframeLinks {
		updated := wpIFrameOpenRX.ReplaceAllString(originalLink, onsVisualOpenTag)
		updated = strings.Replace(updated, iFrameCloseTag, onsVisualCloseTag, 1)

		s.Markdown = strings.Replace(s.Markdown, originalLink, updated, 1)
	}
}

func (s *MarkdownSection) fixFootnotes() {
	fixedFootnotes := make([]string, 0)
	wpFootnotes := wpFootnotesRX.FindAllString(s.Markdown, -1)

	if len(wpFootnotes) > 0 {

		for i, wpFootnote := range wpFootnotes {
			// extract the footnote content from the tags
			florenceFootnote := strings.Replace(wpFootnote, footnotesOpenTag, "", 1)
			florenceFootnote = strings.Replace(florenceFootnote, footnotesCloseTag, "", 1)
			// replace the tags with the index of the footnote
			s.Markdown = strings.Replace(s.Markdown, wpFootnote, fmt.Sprintf(onsFootnoteIndex, i+1), 1)
			fixedFootnotes = append(fixedFootnotes, florenceFootnote)
		}

		onsFootnotes := onsFootnotesTitle
		for i, fn := range fixedFootnotes {
			onsFootnotes += fmt.Sprintf(onsFootnote, i+1, fn)
		}

		s.Markdown += onsFootnotes
	}
}

func (s *MarkdownSection) fixExplanations() {
	explanations := explanationRX.FindAllString(s.Markdown, -1)
	for _, wpExplanation := range explanations {
		onsPulloutBox := explanationOpenRX.ReplaceAllString(wpExplanation, onsPulloutBoxOpenTag)
		onsPulloutBox = strings.Replace(onsPulloutBox, "\"]", onsPulloutBoxCloseTag, 1)
		s.Markdown = strings.Replace(s.Markdown, wpExplanation, onsPulloutBox, 1)
	}
}

func trimTrailingWhiteSpace(body string) string {
	return strings.TrimRightFunc(body, func(c rune) bool {
		return unicode.IsSpace(c)
	})
}

func getHref(t html.Token) string {
	for _, v := range t.Attr {
		if v.Key == hrefTag {
			return v.Val
		}
	}
	return ""
}
