package swan

import (
	"regexp"
	"strings"

	"github.com/andybalholm/cascadia"
)

type extractTitle struct{}

var (
	titleSplitters  = "|-»:"
	titleMatcher    = cascadia.MustCompile("title")
	headlineMatcher = cascadia.MustCompile("meta[name=headline]")
)

func cleanTitle(a *Article, t string) string {
	if sn, ok := a.Meta.OpenGraph["site_name"]; ok {
		t = strings.TrimSpace(strings.Replace(t, sn, "", -1))
	}

	if a.Meta.Domain != "" {
		r, err := regexp.Compile(a.Meta.Domain)
		if err == nil {
			t = strings.TrimSpace(r.ReplaceAllString(t, ""))
		}
	}

	return strings.TrimSpace(strings.Trim(t, titleSplitters))
}

func (e extractTitle) run(a *Article) error {
	title, ok := a.Meta.OpenGraph["title"]

	if !ok {
		title, ok = a.Doc.FindMatcher(headlineMatcher).Attr("content")
	}

	if !ok {
		title = a.Doc.FindMatcher(titleMatcher).Text()
	}

	a.Meta.Title = cleanTitle(a, title)

	return nil
}
