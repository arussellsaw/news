package swan

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

type extractPublishDate struct{}

type publishDate struct {
	m    goquery.Matcher
	attr string
}

var (
	publishDaters = []publishDate{
		publishDate{
			m:    cascadia.MustCompile("[property=rnews\\:datePublished]"),
			attr: "content",
		},
		publishDate{
			m:    cascadia.MustCompile("[property=article\\:published_time]"),
			attr: "content",
		},
		publishDate{
			m:    cascadia.MustCompile("[name=OriginalPublicationDate]"),
			attr: "content",
		},
		publishDate{
			m:    cascadia.MustCompile("[itemprop=datePublished]"),
			attr: "datetime",
		},
	}
)

func (e extractPublishDate) run(a *Article) error {
	for _, pd := range publishDaters {
		s := a.Doc.FindMatcher(pd.m)
		if s.Size() == 0 {
			continue
		}

		t, exists := s.Attr(pd.attr)
		if !exists {
			continue
		}

		a.Meta.PublishDate = t
		break
	}

	return nil
}
