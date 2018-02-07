package zebedee

import "regexp"

const (
	iFrameTagRXPtn          = "\\[iframe url[ ]*=[ ]*.+]"
	iFrameOpenRXPtn         = "\\[iframe url[ ]*=[ ]*"
	iFrameCloseTag          = "]"
	onsVisualOpenTag        = "<ons-interactive url="
	onsVisualCloseTag       = " full-width=\"true\"/>"
	footnotesRXPtn          = "\\[footnote\\].+?\\[\\/footnote\\]$*"
	footnotesOpenTag        = "[footnote]"
	footnotesCloseTag       = "[/footnote]"
	onsFootnoteIndex        = "^%d^"
	onsFootnotesTitle       = "\n\n###Footnotes:"
	onsFootnote             = "\n%d. %s"
	onsHyperlinkInline      = "[%s][%d]"
	onsHyperlink            = "  [%d]: %s\n"
	OpenATag                = "a"
	explanationRXPtn        = "\\[explanation[ ]*content[ ]*=[ ]*.+?\"[ ]*\\]"
	explanationOpenTagRxPtn = "\\[explanation[ ]*content[ ]*=[ ]*\""
	onsPulloutBoxOpenTag    = "<ons-box align=\"full\">"
	onsPulloutBoxCloseTag   = "</ons-box>"
)

var (
	wpFootnotesRX     = regexp.MustCompile(footnotesRXPtn)
	wpIFrameRX        = regexp.MustCompile(iFrameTagRXPtn)
	wpIFrameOpenRX    = regexp.MustCompile(iFrameOpenRXPtn)
	explanationRX     = regexp.MustCompile(explanationRXPtn)
	explanationOpenRX = regexp.MustCompile(explanationOpenTagRxPtn)

	openPlaceholders = map[string]func(string) string{
		"h1": func(body string) string {
			return body + "[h1]"
		},
		"h2": func(body string) string {
			return body + "[h2]"
		},
		"h3": func(body string) string {
			return body + "[h3]"
		},
		"h4": func(body string) string {
			return body + "[h4]"
		},
		"ul": func(body string) string {
			return body + "[ul-start]"
		},
		"li": func(body string) string {
			body = trimTrailingWhiteSpace(body)
			return body + "[ul-item]"
		},
	}

	closePlaceholders = map[string]func(string) string{}

	onsMarkdown = map[string]string{
		"[ul-start]": "",
		"[ul-item]":  "\n- ",
		"[h1]":       "#",
		"[h2]":       "##",
		"[h3]":       "###",
		"[h4]":       "####",
	}
)
