package zebedee

import (
	"github.com/ONSdigital/dp-visual-ons-migration/migration"
	"fmt"
	"strings"
	"golang.org/x/net/html"
)

type Href struct {
	Index int
	URL   string
	Text  string
	Close bool
}

func ConvertHTMLToONSMarkdown(section string, plan *migration.Plan) (string, error) {
	body := strings.NewReader(section)
	z := html.NewTokenizer(body)
	z.AllowCDATA(true)

	markdownBody := ""
	linkIndex := 0
	links := make([]*Href, 0)

htmlTokenizer:
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			fmt.Println("encountered error " + z.Token().String())
			break htmlTokenizer
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == OpenATag {
				linkIndex += 1

				href := &Href{
					Index: linkIndex,
					URL:   plan.GetMigratedURL(getHref(t)),
					Text:  "",
					Close: false,
				}
				links = append(links, href)

				markdownBody += fmt.Sprintf("[link-%d]", linkIndex)

			} else if t.Data == "img" {
				markdownBody += fmt.Sprintf(imageFormat, getAttr(t, "src"))
			} else if applyPlaceholder, ok := openPlaceholders[t.Data]; ok {
				markdownBody = applyPlaceholder(markdownBody)
			}
		case tt == html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "img" {

				imgVal := fmt.Sprintf(imageFormat, getAttr(t))
				markdownBody = appendToBody(links, markdownBody, imgVal)
			}
		case tt == html.EndTagToken:
			t := z.Token()

			if t.Data == "a" {
				for _, href := range links {
					if !href.Close {
						// close the first open one.
						href.Close = true
					}
				}
			} else if applyPlaceholder, ok := closePlaceholders[t.Data]; ok {
				markdownBody = applyPlaceholder(markdownBody)
			}

		case tt == html.TextToken:
			t := z.Token()
			markdownBody = appendToBody(links, markdownBody, html.UnescapeString(t.String()))
		}
	}

	if len(links) > 0 {
		linksFooter := "\n\n\n"
		for _, link := range links {
			old := fmt.Sprintf("[link-%d]", link.Index)
			new := fmt.Sprintf("[%s][%d]", link.Text, link.Index)
			markdownBody = strings.Replace(markdownBody, old, new, 1)
			linksFooter += fmt.Sprintf(onsHyperlink, link.Index, link.URL)
		}
		markdownBody += linksFooter
	}

	for placeHolder, val := range onsMarkdown {
		markdownBody = strings.Replace(markdownBody, placeHolder, val, -1)
	}

	return markdownBody, nil
}

func appendToBody(links []*Href, body string, value string) string {
	var openHref *Href
	for _, href := range links {
		if !href.Close {
			openHref = href
			break
		}
	}
	if openHref != nil {
		openHref.Text += value
	} else {
		body += value
	}
	return body
}

func getAttr(t html.Token, name string) string {
	for _, v := range t.Attr {
		if v.Key == name {
			return v.Val
		}
	}
	return ""
}
