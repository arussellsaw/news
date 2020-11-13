package swan

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

type extractMetas struct{}
type metaDetectLanguage struct{}

var (
	htmlTagMatcher         = cascadia.MustCompile("html")
	metaOpenGraphMatcher   = cascadia.MustCompile("[property^=og\\:]")
	baseURLMatcher         = cascadia.MustCompile("base[href]")
	metaMatcher            = cascadia.MustCompile("meta")
	metaMatcherCanonical   = cascadia.MustCompile("[name=canonical]")
	metaMatcherDescription = cascadia.MustCompile("[name=description]")
	metaMatcherDomain      = cascadia.MustCompile("[name=domain]")
	metaMatcherFavicon     = cascadia.MustCompile("link[rel~=icon]")
	metaMatcherKeywords    = cascadia.MustCompile("[name=keywords]")
	metaMatcherLangs       = []cascadia.Selector{
		cascadia.MustCompile("[http-equiv=Content-Language]"),
		cascadia.MustCompile("[name=lang]"),
	}
)

func extractMetaLanguage(a *Article, metas *goquery.Selection) {
	lang, _ := a.Doc.FindMatcher(htmlTagMatcher).Attr("lang")

	if lang == "" {
		for _, s := range metaMatcherLangs {
			lang, _ = metas.FilterMatcher(s).Attr("content")
			if lang != "" {
				break
			}
		}
	}

	if lang != "" {
		a.Meta.Lang = lang[:2]
	}
}

func (e extractMetas) run(a *Article) error {
	metas := a.Doc.FindMatcher(metaMatcher)

	extractMetaLanguage(a, metas)

	t, _ := metas.FilterMatcher(metaMatcherCanonical).Attr("content")
	a.Meta.Canonical = strings.TrimSpace(t)

	t, _ = metas.FilterMatcher(metaMatcherDescription).Attr("content")
	a.Meta.Description = strings.TrimSpace(t)

	t, _ = metas.FilterMatcher(metaMatcherDomain).Attr("content")
	a.Meta.Domain = strings.TrimSpace(t)

	t, _ = a.Doc.FindMatcher(metaMatcherFavicon).Attr("href")
	a.Meta.Favicon = strings.TrimSpace(t)

	t, _ = metas.FilterMatcher(metaMatcherKeywords).Attr("content")
	a.Meta.Keywords = strings.TrimSpace(t)

	a.Meta.OpenGraph = make(map[string]string)
	metas.FilterMatcher(metaOpenGraphMatcher).Each(
		func(i int, s *goquery.Selection) {
			if content, exists := s.Attr("content"); exists {
				prop, _ := s.Attr("property")
				a.Meta.OpenGraph[prop[3:]] = content
			}
		})

	t, _ = a.Doc.FilterMatcher(baseURLMatcher).Attr("href")
	if t == "" {
		t = a.URL
	}

	var err error
	a.baseURL, err = url.Parse(t)
	if err != nil {
		a.baseURL, _ = url.Parse(a.URL)
	}

	return nil
}

func (m metaDetectLanguage) run(a *Article) error {
	_, hasLang := stopwords[a.Meta.Lang]

	if a.Meta.Lang == "" || !hasLang {
		a.Meta.Lang = detectLang(a)
	}

	return nil
}
