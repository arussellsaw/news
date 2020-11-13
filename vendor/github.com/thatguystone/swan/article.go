package swan

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Article is a fully extracted and cleaned document.
type Article struct {
	// Final URL after all redirects
	URL string

	// Newline-separated and cleaned content
	CleanedText string

	// Node from which CleanedText was created. Call .Html() on this to get
	// printable HTML.
	TopNode *goquery.Selection

	// A header image to use for the article. Nil if no image could be
	// detected.
	Img *Image

	// All metadata associated with the original document
	Meta struct {
		Authors     []string
		Canonical   string
		Description string
		Domain      string
		Favicon     string
		Keywords    string
		Links       []string
		Lang        string
		OpenGraph   map[string]string
		PublishDate string
		Tags        []string
		Title       string
	}

	// Full document backing this article
	Doc *goquery.Document

	// For use resolving URLs in the document
	baseURL *url.URL

	// Caches information about nodes so that it doesn't have to be updated
	cCache map[*html.Node]*contentCache

	// Scores that have been calculated
	scores map[*html.Node]int
}

// Image contains information about the header image associated with an
// article
type Image struct {
	Src        string
	Width      uint
	Height     uint
	Bytes      int64
	Confidence uint
	Sel        *goquery.Selection
}

type contentCache struct {
	text            string
	wordCount       uint
	stopwords       uint
	highLinkDensity bool
	s               *goquery.Selection
}

type runner interface {
	run(a *Article) error
}

type processor struct {
	probe   func(a *Article) uint
	runners []runner
}

type useKnownArticles struct{}

const (
	imgHeader = `<p class="image-container" style="text-align: center;">` +
		`<a href="%s"><img title="%s" src="%s"/></a>` +
		`</p>`
)

var (
	baseRunners = []runner{
		extractMetas{},

		extractAuthors{},
		extractPublishDate{},
		extractTags{},
		extractTitle{},

		precleanup{},
		metaDetectLanguage{},
		useKnownArticles{},
		cleanup{},
	}

	defaultProcessor = &processor{
		probe: func(a *Article) uint {
			return 1
		},
		runners: []runner{
			extractTopNode{},
			extractLinks{},
			extractImages{},
			extractVideos{},

			// Does more document mangling and TopNode resetting
			extractContent{},
		},
	}

	processors = []*processor{
		comicProcessor,

		// Always last since it's the default
		defaultProcessor,
	}

	// Don't match all-at-once: there's precedence here
	knownArticles = []goquery.Matcher{
		cascadia.MustCompile("[itemprop=articleBody]"),
		cascadia.MustCompile("[itemprop=blogPost]"),
		cascadia.MustCompile(".post-content"),
		cascadia.MustCompile("article"),
	}
)

func (u useKnownArticles) run(a *Article) error {
	for _, m := range knownArticles {
		for _, n := range a.Doc.FindMatcher(m).Nodes {
			cc := a.getCCache(n)

			// Sometimes even "known" articles are wrong
			if cc.stopwords > 5 && !cc.highLinkDensity {
				// Remove from document so that memory can be freed
				if n.Parent != nil {
					n.Parent.RemoveChild(n)
				}

				a.Doc = goquery.NewDocumentFromNode(n)
				a.TopNode = a.Doc.Selection
				return nil
			}
		}
	}

	return nil
}

// Checks to see if TopNode is a known article tag that was picked before
// scoring
func (u useKnownArticles) isKnownArticle(a *Article) bool {
	for _, m := range knownArticles {
		if a.Doc.IsMatcher(m) {
			return true
		}
	}

	return false
}

func (a *Article) extract() error {
	var p *processor

	a.cCache = make(map[*html.Node]*contentCache)
	a.scores = make(map[*html.Node]int)

	for _, r := range baseRunners {
		err := r.run(a)
		if err != nil {
			return err
		}
	}

	max := uint(0)
	for _, pp := range processors {
		score := pp.probe(a)
		if score > max {
			p = pp
			max = score
		}
	}

	for _, r := range p.runners {
		err := r.run(a)
		if err != nil {
			return err
		}
	}

	a.cCache = nil
	a.scores = nil
	a.baseURL = nil

	return nil
}

func (a *Article) getCCache(n *html.Node) *contentCache {
	cc, ok := a.cCache[n]
	if !ok {
		s := goquery.NewDocumentFromNode(n).Selection
		cc = &contentCache{
			text: strings.TrimSpace(s.Text()),
			s:    s,
		}

		ws := splitText(cc.text)
		cc.wordCount = uint(len(ws))
		cc.stopwords = stopwordCountWs(a.Meta.Lang, ws)
		cc.highLinkDensity = highLinkDensity(cc)
		a.cCache[n] = cc
	}

	return cc
}

func highLinkDensity(cc *contentCache) bool {
	var b bytes.Buffer

	links := cc.s.FindMatcher(linkTags)

	if links.Size() == 0 {
		return false
	}

	links.Each(func(i int, l *goquery.Selection) {
		b.WriteString(l.Text())
	})

	linkWords := float32(strings.Count(b.String(), " "))

	return ((linkWords / float32(cc.wordCount)) * float32(len(cc.s.Nodes))) >= 1
}

func (a *Article) addInlineArticleImageHTML(title string) {
	if a.Img == nil {
		return
	}

	if a.TopNode == nil {
		a.TopNode = goquery.NewDocumentFromNode(&html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Span,
			Data:     "span",
		}).Selection
	}

	a.TopNode.PrependHtml(fmt.Sprintf(imgHeader,
		html.EscapeString(a.URL),
		html.EscapeString(title),
		html.EscapeString(a.Img.Src)))
}
