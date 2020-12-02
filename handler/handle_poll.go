package handler

import (
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
	"net/http"
)

func handlePoll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	urls := make(map[string]struct{})
	sources, err := dao.GetAllSources(ctx)
	if err != nil {
		slog.Error(ctx, "Error getting sources: %s", err)
		return
	}
	for _, s := range sources {
		if _, ok := urls[s.FeedURL]; ok {
			continue
		}
		urls[s.FeedURL] = struct{}{}
		err := p.Publish(ctx, "sources", SourceEvent{Source: domain.Source{FeedURL: s.FeedURL}})
		if err != nil {
			httpError(ctx, w, "Error marshaling pubsub event", err)
			return
		}
		slog.Info(ctx, "Dispatched source %s", s.FeedURL)
	}
}
