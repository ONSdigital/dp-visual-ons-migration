package zebedee

import "regexp"

const (
	iFrameTagRXPtn     = "\\[iframe url[ ]*=[ ]*.+]"
	iFrameOpenRXPtn    = "\\[iframe url[ ]*=[ ]*"
	iFrameCloseTag     = "]"
	onsVisualOpenTag   = "<ons-interactive url="
	onsVisualCloseTag  = " full-width=\"true\"/>"
	footnotesRXPtn     = "\\[footnote\\].+?\\[\\/footnote\\]$*"
	footnotesOpenTag   = "[footnote]"
	footnotesCloseTag  = "[/footnote]"
	onsFootnoteIndex   = "^%d^"
	onsFootnotesTitle  = "\n\n###Footnotes:"
	onsFootnote        = "\n%d. %s"
	onsHyperlinkInline = "[%s][%d]"
	onsHyperlink       = "  [%d]: %s\n"
	OpenATag           = "a"
)

var (
	wpFootnotesRX  = regexp.MustCompile(footnotesRXPtn)
	wpIFrameRX     = regexp.MustCompile(iFrameTagRXPtn)
	wpIFrameOpenRX = regexp.MustCompile(iFrameOpenRXPtn)

	htmlMarkdownMapping = map[string]string{
		"h1": "#",
		"h2": "##",
		"h3": "###",
		"h4": "####",
		"h5": "#####",
	}
)
