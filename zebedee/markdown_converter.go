package zebedee

import (
	"fmt"
	"strings"
	"golang.org/x/net/html"
)

func convertHTMLToONSMarkdown(section string) (string, error) {
	body := strings.NewReader(section)
	z := html.NewTokenizer(body)
	z.AllowCDATA(true)

	markdownBody := ""
	linkIndex := 1
	links := make([]string, 0)

parser:
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			fmt.Println("encountered error " + z.Token().String())
			break parser

		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == OpenATag {
				links = append(links, getHref(t))
				linkBody := ""

			findClosingATag:
				for {
					tt = z.Next()

					switch {
					case tt == html.TextToken:
						t := z.Token()
						linkBody += html.UnescapeString(t.String())
					case tt == html.EndTagToken:
						t = z.Token()
						linkIndex += 1
						break findClosingATag
					case tt == html.ErrorToken:
						break findClosingATag
					}
				}

				markdownBody += fmt.Sprintf(onsHyperlinkInline, linkBody, linkIndex)

			} else if markdown, ok := htmlMarkdownMapping[t.Data]; ok {
				markdownBody += markdown
			}
		case tt == html.TextToken:
			t := z.Token()

			markdownBody += html.UnescapeString(t.String())
		}
	}

	// now append the links to the bottom of the article in the ONS florence format
	if len(links) > 0 {
		markdownBody += "\n\n\n"
		for i, link := range links {
			markdownBody += fmt.Sprintf(onsHyperlink, i+1, link)
		}
	}

	return markdownBody, nil
}

func getHref(t html.Token) string {
	for _, v := range t.Attr {
		if v.Key == "href" {
			return v.Val
		}
	}
	return ""
}
