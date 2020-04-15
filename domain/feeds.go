package domain

import (
	"context"
	"github.com/monzo/slog"
	"sort"
	"strings"
	"sync"

	"github.com/arussellsaw/news/pkg/goose"
	"github.com/mmcdole/gofeed"
	"golang.org/x/sync/errgroup"
)

var (
	amu      sync.RWMutex
	articles []Article
	links    map[string]Article
)

func GetArticleForLink(link string) (Article, bool) {
	amu.RLock()
	defer amu.RLock()
	a, ok := links[link]
	return a, ok
}

func GetArticles(ctx context.Context) ([]Article, error) {
	amu.RLock()
	if len(articles) != 0 {
		return articles, nil
	}
	amu.RUnlock()

	amu.Lock()
	defer amu.Unlock()
	if len(articles) != 0 {
		return articles, nil
	}
	aa, err := doGetArticles(ctx)
	if err != nil {
		return nil, nil
	}
	articles = aa
	links = make(map[string]Article)
	for _, article := range articles {
		links[article.Link] = article
	}
	return articles, nil
}

func doGetArticles(ctx context.Context) ([]Article, error) {
	eg := errgroup.Group{}
	articles := make(chan Article, 1024^2)
	bucket := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		bucket <- struct{}{}
	}
	for _, s := range sources {
		s := s
		eg.Go(func() error {
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(s.FeedURL)
			if err != nil {
				return err
			}

			g := errgroup.Group{}
			for _, item := range feed.Items {
				item := item
				s := s
				<-bucket
				g.Go(func() error {
					defer func() { bucket <- struct{}{} }()
					var imageURL string
					if item.Image != nil {
						imageURL = item.Image.URL
					}
					slogParams := map[string]string{
						"url":       item.Link,
						"image_url": imageURL,
					}
					var content string
					text, image, ok, err := cache.Get(item.Link)
					if err != nil {
						return err
					}
					if ok {
						content = text
						imageURL = image
					} else {
						g := goose.New()
						article, err := g.ExtractFromURL(item.Link)
						if err != nil {
							slog.Error(ctx, "Error fetching article: %s", err, slogParams)
							return nil
						}
						content = article.CleanedText
						imageURL = article.TopImage
						err = cache.Set(item.Link, content, imageURL)
						if err != nil {
							return err
						}
					}

					var author string
					if item.Author != nil {
						author = item.Author.Name
					}

					articles <- Article{
						Title:       item.Title,
						Description: item.Description,
						Content:     toElements(content, "\n"),
						ImageURL:    imageURL,
						Link:        item.Link,
						Author:      author,
						Source:      s,
						Timestamp:   *item.PublishedParsed,
						TS:          item.PublishedParsed.Format("15:04 02-01-2006"),
					}
					return nil
				})
			}
			return g.Wait()
		})
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	close(articles)
	out := []Article{}
	for a := range articles {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp.Before(out[j].Timestamp)
	})
	return out, nil
}

func matchClass(class string, exclude []string) bool {
	for _, e := range exclude {
		if strings.Contains(class, e) {
			return true
		}
	}
	return false
}

func FilterArticles(aa []Article, cat string) []Article {
	out := []Article{}
L:
	for _, a := range aa {
		for _, c := range a.Source.Categories {
			if c == cat || cat == "" {
				out = append(out, a)
				continue L
			}
		}
	}
	return out
}

func toElements(s, br string) []Element {
	var (
		lines = strings.Split(s, br)
		out   []Element
		row   string
	)
	for _, line := range lines {
		if line == "p" {
			continue
		}
		if line == "" {
			out = append(out, Element{Type: "text", Value: row})
			row = ""
			continue
		}
		row += line

	}
	if row != "" {
		out = append(out, Element{Type: "text", Value: row})
	}
	return out
}
