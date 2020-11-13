package domain

import (
	"context"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/monzo/slog"

	"github.com/mmcdole/gofeed"
	"github.com/thatguystone/swan"
	"golang.org/x/sync/errgroup"
)

var (
	space  = uuid.MustParse("45e990eb-e8d4-4a13-8e74-e544bd11e45d")
	jar, _ = cookiejar.New(&cookiejar.Options{})
	c      = http.Client{
		Jar: jar, //binks
	}
)

func FetchArticles(ctx context.Context) ([]Article, error) {
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
					if !s.DisableFetch {
						res, err := c.Get(item.Link)
						if err != nil {
							slog.Error(ctx, "Error fetching article: %s", err, slogParams)
							return nil
						}
						buf, err := ioutil.ReadAll(res.Body)
						if err != nil {
							slog.Error(ctx, "Error reading article: %s", err, slogParams)
							return nil
						}
						article, err := swan.FromHTML(item.Link, buf)
						if err != nil {
							slog.Error(ctx, "Error parsing article: %s", err, slogParams)
							return nil
						}
						content = article.CleanedText
						if article.Img != nil {
							imageURL = article.Img.Src
						}
						if content == "" {
							content = item.Content
						}
					}

					var author string
					if item.Author != nil {
						author = item.Author.Name
					}

					articles <- Article{
						ID:          uuid.NewHash(fnv.New32(), space, []byte(item.Link), 4).String(),
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
