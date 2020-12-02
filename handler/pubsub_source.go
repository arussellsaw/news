package handler

import (
	"encoding/json"
	"github.com/arussellsaw/news/dao"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/monzo/slog"

	"github.com/arussellsaw/news/domain"
)

var Prefix string

type SourceEvent struct {
	Source domain.Source
}

func handlePubsubSource(w http.ResponseWriter, r *http.Request) {
	var (
		m   domain.PubSubMessage
		e   SourceEvent
		ctx = r.Context()
	)
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		httpError(ctx, w, "Error decoding event", err)
		return
	}

	if err := json.Unmarshal(m.Message.Data, &e); err != nil {
		httpError(ctx, w, "Error decoding feed", err)
		return
	}

	slog.Info(ctx, "handling %s", e.Source.Name)
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(e.Source.FeedURL)
	if err != nil {
		httpError(ctx, w, "Error parsing rss feed", err)
		return
	}

	for _, item := range feed.Items {
		a, err := dao.GetArticleByURL(ctx, item.Link)
		if err != nil {
			slog.Error(ctx, "Error getting article: %s", err)
			continue
		}
		if a != nil {
			continue
		}
		err = p.Publish(ctx, "articles", ArticleEvent{
			Article: domain.Article{
				Title:       strings.TrimSpace(item.Title),
				Description: item.Description,
				ImageURL: func() string {
					if item.Image != nil {
						return item.Image.URL
					}
					return ""
				}(),
				Link:   item.Link,
				Source: e.Source,
				Timestamp: func() time.Time {
					if item.PublishedParsed != nil {
						return *item.PublishedParsed
					}
					return time.Now().Add(-time.Duration(rand.Intn(60)) * time.Minute)
				}(),
			},
		})
		if err != nil {
			httpError(ctx, w, "Error marshaling pubsub event", err)
			return
		}
		slog.Info(ctx, "Dispatched article %s: %s", item.Link, item.Title)
	}
}
