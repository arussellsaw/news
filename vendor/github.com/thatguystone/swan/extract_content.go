package swan

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/tdewolff/minify/v2"
	minhtml "github.com/tdewolff/minify/v2/html"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type extractContent struct{}

var (
	allTags                 = cascadia.MustCompile("*")
	bodyTag                 = cascadia.MustCompile("body")
	pTags                   = cascadia.MustCompile("p")
	replaceWithContentsTags = cascadia.MustCompile("a, b, strong, i, sup")
	goodContent             = cascadia.MustCompile("object, embed, img")

	multiNewlines = []byte("\n\n\n")
	dblNewlines   = []byte("\n\n")
)

func (e extractContent) run(a *Article) error {
	if a.TopNode == nil {
		return nil
	}

	known := useKnownArticles{}
	if !known.isKnownArticle(a) {
		e.addSiblings(a)
	}

	// This is the equivalent of python-goose's post_cleanup
	a.TopNode.Children().FilterFunction(func(i int, s *goquery.Selection) bool {
		// Center nodes in an article? Get real.
		if nodeIs(s.Nodes[0], atom.Center) {
			return true
		}

		if !nodeIs(s.Nodes[0], atom.P) {
			cc := a.getCCache(s.Nodes[0])
			return cc.highLinkDensity ||
				e.noParasWithoutTable(s) ||
				!e.isNodeScoreThreshMet(a, s)
		}

		return false
	}).Remove()

	e.dropNegativeScored(a)
	e.dropTinyEls(a)
	e.prepareHTMLOut(a)
	e.prepareCleanedText(a)

	return nil
}

func (e extractContent) prepareHTMLOut(a *Article) error {
	if a.TopNode == nil {
		return nil
	}

	m := minify.New()
	m.AddFunc("text/html", minhtml.Minify)

	var b bytes.Buffer
	html, _ := a.TopNode.Html()

	m.Minify("text/html", &b, strings.NewReader(html))
	doc, _ := goquery.NewDocumentFromReader(&b)

	a.TopNode = doc.FindMatcher(bodyTag)

	// Quick-and-dirty node-to-text replacement
	a.TopNode.FindMatcher(replaceWithContentsTags).Each(
		func(i int, s *goquery.Selection) {
			s.Contents().Unwrap()
		})

	a.addInlineArticleImageHTML(a.Meta.Title)

	return nil
}

func (e extractContent) prepareCleanedText(a *Article) {
	buff := bytes.Buffer{}
	rplc := strings.NewReplacer("\n", " ", "\r", "")

	var textify func(n *html.Node)
	textify = func(n *html.Node) {
		switch {
		case n.Type == html.TextNode:
			// TODO(astone): what about text in textareas?
			rplc.WriteString(&buff, n.Data)

		case n.Type == html.ElementNode:
			switch n.DataAtom {
			case atom.Br:
				buff.WriteRune('\n')
			case atom.P:
				buff.WriteString("\n\n")
			}
			fallthrough

		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				textify(c)
			}
		}
	}

	for _, n := range a.TopNode.Nodes {
		textify(n)
	}

	// Get rid of large gaps
	b := buff.Bytes()
	l := len(b) + 1
	for len(b) != l {
		l = len(b)
		b = bytes.Replace(b, multiNewlines, dblNewlines, -1)
	}

	a.CleanedText = string(bytes.TrimSpace(b))
}

func (e extractContent) addSiblings(a *Article) {
	baseScore := e.getSiblingBaseScore(a)

	a.TopNode.PrevAll().Each(func(i int, s *goquery.Selection) {
		a.TopNode.PrependNodes(e.getSiblingContent(a, s, baseScore)...)
	})
}

func (e extractContent) getSiblingContent(
	a *Article,
	s *goquery.Selection,
	baseScore uint) []*html.Node {

	var ret []*html.Node

	if nodeIs(s.Nodes[0], atom.P) && len(s.Text()) > 0 {
		return s.Nodes
	}

	ps := s.FindMatcher(pTags)
	for _, n := range ps.Nodes {
		cc := a.getCCache(n)
		if len(cc.text) > 0 {
			if cc.stopwords > baseScore && !cc.highLinkDensity {
				ret = append(ret, createNode(atom.P, "p", cc.text))
			}
		}
	}

	return ret
}

func (e extractContent) getSiblingBaseScore(a *Article) uint {
	base := uint(100000)
	pCount := uint(0)
	pScore := uint(0)

	for _, n := range a.TopNode.FindMatcher(pTags).Nodes {
		cc := a.getCCache(n)

		if cc.stopwords > 2 && !cc.highLinkDensity {
			pCount++
			pScore += cc.stopwords
		}
	}

	if pCount > 0 {
		base = pScore / pCount
	}

	base = uint(float32(base) * float32(0.3))

	return base
}

func (e extractContent) noParasWithoutTable(s *goquery.Selection) bool {
	s.FindMatcher(pTags).Each(func(i int, s *goquery.Selection) {
		if len(s.Text()) < 25 {
			s.Remove()
		}
	})

	return s.FindMatcher(pTags).Length() == 0 && !nodeIs(s.Nodes[0], atom.Td)
}

func (e extractContent) isNodeScoreThreshMet(a *Article, s *goquery.Selection) bool {
	topNodeScore := a.scores[a.TopNode.Nodes[0]]
	currNodeScore := a.scores[s.Nodes[0]]
	threshScore := int(float32(topNodeScore) * 0.08)

	if (currNodeScore < threshScore) && !nodeIs(s.Nodes[0], atom.Td) {
		return false
	}

	return true
}

func (e extractContent) dropTinyEls(a *Article) {
	a.TopNode.Children().Each(func(i int, s *goquery.Selection) {
		cc := a.getCCache(s.Nodes[0])
		remove := s.HasMatcher(goodContent).Length() == 0 &&
			((cc.stopwords < 3 || cc.highLinkDensity) ||
				(strings.HasPrefix(cc.text, "(") &&
					strings.HasSuffix(cc.text, ")")))

		if remove {
			s.Remove()
		}
	})
}

func (e extractContent) dropNegativeScored(a *Article) {
	for n, score := range a.scores {
		if score <= 0 {
			if n.Parent != nil {
				n.Parent.RemoveChild(n)
			}
		}
	}
}
