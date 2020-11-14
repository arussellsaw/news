package handler

import (
	"encoding/json"
	"net/http"

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

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(e.Source.FeedURL)
	if err != nil {
		httpError(ctx, w, "Error parsing rss feed", err)
		return
	}

	for _, item := range feed.Items {
		err := p.Publish(ctx, "articles", ArticleEvent{
			Article: domain.Article{
				Title:       item.Title,
				Description: item.Description,
				ImageURL: func() string {
					if item.Image != nil {
						return item.Image.URL
					}
					return ""
				}(),
				Link:   item.Link,
				Source: e.Source,
			},
		})
		if err != nil {
			httpError(ctx, w, "Error marshaling pubsub event", err)
			return
		}
		slog.Info(ctx, "Dispatched article %s: %s", item.Link, item.Title)
	}
}
