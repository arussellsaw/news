package swan

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
)

const (
	textChunkLen = 8192
)

var (
	unsupportedLangs = map[string]bool{
		// According to http://en.wikipedia.org/wiki/List_of_ISO_639-1_codes,
		// NB is covered by macrolanguage NO, so don't use.
		"nb": true,

		"zh": true, // Needs a good segmenter
	}
)

func init() {
	for k := range unsupportedLangs {
		delete(stopwords, k)
	}
}

func getArticleTextChunk(a *Article) string {
	var b bytes.Buffer
	var getText func(n *html.Node)

	getText = func(p *html.Node) {
		if p.Type == html.TextNode {
			b.WriteString(strings.TrimSpace(p.Data))
			b.WriteByte(' ')
		} else if p.FirstChild != nil {
			for n := p.FirstChild; n != nil; n = n.NextSibling {
				getText(n)
				if b.Len() >= textChunkLen {
					return
				}
			}
		}
	}

	for _, n := range a.Doc.Nodes {
		getText(n)
		if b.Len() >= textChunkLen {
			break
		}
	}

	return b.String()
}

func splitText(t string) (ws []string) {
	start := 0
	inWord := false

	for i, r := range t {
		sep := unicode.IsPunct(r) || unicode.IsSpace(r)

		if sep {
			switch {
			case r == '\'': // Accept things like "boy's"

			case inWord:
				ws = append(ws, t[start:i])
				start = i + 1
				inWord = false

			default:
				start += utf8.RuneLen(r)
			}
		}

		inWord = !sep
	}

	if start < len(t) {
		ws = append(ws, t[start:])
	}

	return
}

func detectLang(a *Article) string {
	score := uint(0)
	detected := "en"
	ws := splitText(getArticleTextChunk(a))

	for lang := range stopwords {
		count := stopwordCountWs(lang, ws)
		if count > score {
			detected = lang
			score = count
		}
	}

	return detected
}

func stopwordCount(lang string, text string) uint {
	ws := splitText(text)
	return stopwordCountWs(lang, ws)
}

func stopwordCountWs(lang string, ws []string) uint {
	words := stopwords[lang]

	count := uint(0)
	for _, w := range ws {
		if _, ok := words[strings.ToLower(w)]; ok {
			count++
		}
	}

	return count
}
