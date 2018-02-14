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
	PostTitle    string   `json:"postTitle"`
	TaxonomyURI  string   `json:"taxonomyURI"`
	RelatedLinks []string `json:"links"`
	Keywords     []string `json:"keywords"`
	VisualURL    string   `json:"visualURL"`
}

// Top level structure holding all the migration details.
type Plan struct {
	VisualExport        *VisualExport
	Mapping             *Mapping
	NationalArchivesURL string
}

// mapping of the posts to migrate - from -> to.
type Mapping struct {
	ToMigrate []*Article
	NotToMigrated map[string]*Article
}

type Attachment struct {
	Title string
	URL   *url.URL
	ID    string
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

func (m *Mapping) GetArticleByURL(target string) (*Article, bool) {
	for _, a := range m.ToMigrate {
		if target == a.VisualURL {
			return a, true
		}
	}
	return nil, false
}

func (p *Plan) GetMigratedURL(current string) string {
	currentURL, _ := url.Parse(current)
	data := log.Data{"url": current}

	if currentURL.Host == "visual.ons.gov.uk" {

		// check if the url is a migrated visual attachment - if so return the url for its migrated location.
		if attachment, ok := p.VisualExport.Attachments[current]; ok {
			log.Debug("visual attachment url found", data)
			return staticONSHost + strings.Replace(attachment.URL.Path, wpAttachmentPath, staticONSPath, 1)
		}

		// otherwise check if the url is a migrated visual post then return the URL of where the post will be migrated to
		if migrationPost, ok := p.Mapping.GetArticleByURL(current); ok {
			log.Debug("visual migration post url found", data)
			return ONSSite + migrationPost.TaxonomyURI
		}

		log.Debug("visual url found but not attachment or migration post", data)
		return p.NationalArchivesURL + current
	}

	// its not a visual post or attachment - so no transformation required nothing.
	return current
}

// add an attachment to the visual mapping
func (m *VisualExport) addAttachment(i *gofeed.Item) error {
	attachmentURL, err := url.Parse(i.Extensions["wp"]["attachment_url"][0].Value)
	if err != nil {
		return Error{"failed to parse visual attachment URL", err, log.Data{"title": i.Title, "url": i.Link}}
	}

	postID := i.Extensions["wp"]["post_id"][0].Value
	m.Attachments[attachmentURL.String()] = &Attachment{URL: attachmentURL, Title: i.Title, ID: postID}
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

func (m *VisualExport) GetThumbNailURL(thumbnailID string) *url.URL {
	for _, attachment := range m.Attachments {
		if attachment.ID == thumbnailID {
			return attachment.URL
		}
	}
	return nil
}

func (a *Article) Valid() error {
	if a.PostTitle == "" {
		return Error{Message: "invalid mapping article title is empty", OriginalErr: nil, Params: nil}
	}
	if a.TaxonomyURI == "" {
		return Error{Message: "invalid mapping article taxonomy uri is empty", OriginalErr: nil, Params: nil}
	}
	if a.VisualURL == "" {
		return Error{Message: "invalid mapping article visual url is empty", OriginalErr: nil, Params: nil}
	}
	return nil
}

func (e Error) Error() string {
	msg := ""
	if e.Message != "" {
		msg += e.Message + " "
	}
	if e.OriginalErr != nil {
		msg += e.OriginalErr.Error()
	}
	return msg
}

func (m *Article) GetTaxonomyURI() string {
	return m.TaxonomyURI + "/" + strings.TrimSpace(strings.ToLower(m.GetCollectionName()))
}

func (m *Article) GetCollectionName() string {
	r, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(r.ReplaceAllString(m.PostTitle, ""))
}
