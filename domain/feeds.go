package domain

import (
	"context"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/arussellsaw/news/pkg/goose"
	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
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
				g.Go(func() error {
					var imageURL string
					if item.Image != nil {
						imageURL = item.Image.URL
					}
					var content string
					if (len(item.Content) < 100 || s.ForceFetch) && !s.DisableFetch {
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
								cache.Set(item.Link, "error fetching content.", "")
								log.Print(err)
								return nil
							}
							content = article.CleanedText
							imageURL = article.TopImage
							err = cache.Set(item.Link, content, imageURL)
							if err != nil {
								return err
							}
						}
					} else {
						doc, err := html.Parse(strings.NewReader(item.Content))
						if err != nil {
							return err
						}

						var f func(n *html.Node)
						f = func(n *html.Node) {
							switch n.Type {
							case html.ElementNode:
								if n.Data == "img" {
									for _, a := range n.Attr {
										if a.Key == "src" {
											if imageURL == "" {
												imageURL = a.Val
											}
										}
									}
								}
							case html.TextNode:
								if n.Parent.Data == "p" {
									attrs := n.Parent.Attr
									for _, a := range attrs {
										if a.Key == "class" && matchClass(a.Val, []string{"twite", "top-stories"}) {
											goto recurse
										}
									}
									if item.Description == "" {
										item.Description = n.Data
									}
									content += n.Data + " "
								}
							}
						recurse:
							for c := n.FirstChild; c != nil; c = c.NextSibling {
								f(c)
							}
						}
						f(doc)
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
