package swan

import "github.com/PuerkitoBio/goquery"

type extractLinks struct{}

func (e extractLinks) run(a *Article) error {
	if a.TopNode == nil {
		return nil
	}

	a.TopNode.FindMatcher(linkTags).Each(func(i int, s *goquery.Selection) {
		h, exists := s.Attr("href")
		if exists && h != "" {
			a.Meta.Links = append(a.Meta.Links, h)
		}
	})

	return nil
}
