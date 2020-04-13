package main

import (
	"github.com/arussellsaw/news/pkg/goose"
	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
	"golang.org/x/sync/errgroup"
	"sort"
	"strings"
	"time"
)

type Source struct {
	Name    string
	URL     string
	FeedURL string
}

type Article struct {
	Title       string
	Description string
	Content     string
	ImageURL    string
	Link        string
	Author      string
	Source      Source
	Timestamp   time.Time
	TS          string
	Layout      Layout
}

var sources = []Source{
	{
		Name:    "Vox",
		URL:     "https://vox.com",
		FeedURL: "https://www.vox.com/rss/index.xml",
	},
	{
		Name:    "The Verge",
		URL:     "https://theverge.com",
		FeedURL: "https://www.theverge.com/rss/index.xml",
	},
	{
		Name:    "Polygon",
		URL:     "https://polygon.com",
		FeedURL: "https://www.polygon.com/rss/index.xml",
	},
	{
		Name:    "TechCrunch",
		URL:     "https://techcrunch.com",
		FeedURL: "http://feeds.feedburner.com/TechCrunch/",
	},
	{
		Name:    "BBC News",
		URL:     "https://bbc.co.uk/news",
		FeedURL: "http://feeds.bbci.co.uk/news/rss.xml",
	},
	{
		Name:    "The Guardian",
		URL:     "https://theguardian.co.uk",
		FeedURL: "https://www.theguardian.com/uk/rss",
	},
}

func getArticles() ([]Article, error) {
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
				g.Go(func() error {
					var imageURL string
					if item.Image != nil {
						imageURL = item.Image.URL
					}
					var content string
					if len(item.Content) < 100 {
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
								return err
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
						Content:     content,
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
