package migration

import (
	"github.com/ONSdigital/go-ns/log"
	"regexp"
	"strings"
	"github.com/mmcdole/gofeed"
	"net/url"
)

type Error struct {
	Message     string
	OriginalErr error
	Params      log.Data
}

// Article a florence article
type Article struct {
	PostTitle   string   `json:"postTitle"`
	TaxonomyURI string   `json:"taxonomyURI"`
	Keywords    []string `json:"keywords"`
	VisualURL   string   `json:"visualURL"`
}

// Top level structure holding all the migration details.
type Plan struct {
	VisualExport *VisualExport
	Mapping      *Mapping
}

// mapping of the posts to migrate - from -> to.
type Mapping struct {
	PostsToMigrate map[string]*Article
}

type Attachment struct {
	Title string
	URL   *url.URL
}

type VisualExport struct {
	Attachments map[string]*Attachment
	Posts       map[string]*gofeed.Item
}

func newVisualExport() *VisualExport {
	return &VisualExport{
		Attachments: make(map[string]*Attachment),
		Posts:       make(map[string]*gofeed.Item),
	}
}

func (p *Plan) GetMigratedURL(current string) string {
	// check if the url is a migrated visual attachment - if so return the url for its migrated location.
	if attachment, ok := p.VisualExport.Attachments[current]; ok {
		return staticONSHost + strings.Replace(attachment.URL.Path, wpAttachmentPath, staticONSPath, 1)
	}

	// otherwise check if the url is a migrated visual post then return the URL of where the post will be migrated to
	for _, post := range p.Mapping.PostsToMigrate {
		if post.VisualURL == current {
			return post.TaxonomyURI
		}
	}

	// its not a visual link - so do nothing.
	return current
}

// add an attachment to the visual mapping
func (m *VisualExport) addAttachment(i *gofeed.Item) error {
	attachmentURL, err := url.Parse(i.Extensions["wp"]["attachment_url"][0].Value)
	if err != nil {
		return Error{"failed to parse visual attachment URL", err, log.Data{"title": i.Title, "url": i.Link}}
	}

	m.Attachments[attachmentURL.String()] = &Attachment{URL: attachmentURL, Title: i.Title}
	return nil
}

// add a post to the visual mapping
func (m *VisualExport) addPost(i *gofeed.Item) error {
	postURL, err := url.Parse(i.Link)
	if err != nil {
		return Error{"failed to parse visual URL", err, log.Data{"title": i.Title, "url": i.Link}}
	}

	if _, ok := m.Posts[postURL.String()]; !ok {
		m.Posts[postURL.String()] = i
	} else {
		return Error{"duplicate entry in visual RSS xmL", err, log.Data{"title": i.Title, "url": i.Link}}
	}
	return nil
}

func (e Error) Error() string {
	return e.Message + ": " + e.OriginalErr.Error()
}

func (m *Article) GetTaxonomyURI() string {
	return m.TaxonomyURI + "/" + strings.TrimSpace(strings.ToLower(m.GetCollectionName()))
}

func (m *Article) GetCollectionName() string {
	r, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(r.ReplaceAllString(m.PostTitle, ""))
}
