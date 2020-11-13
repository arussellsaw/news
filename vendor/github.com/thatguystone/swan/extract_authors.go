package swan

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

type extractAuthors struct{}

var (
	authorMatcher = cascadia.MustCompile("[itemprop~=author] [itemprop=name]")
)

func (e extractAuthors) run(a *Article) error {
	auths := make(map[string]interface{})

	a.Doc.FindMatcher(authorMatcher).Each(func(i int, s *goquery.Selection) {
		t := s.Text()
		if t != "" {
			auths[strings.TrimSpace(t)] = nil
		}
	})

	for k := range auths {
		a.Meta.Authors = append(a.Meta.Authors, k)
	}

	return nil
}
