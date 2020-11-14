package handler

import (
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
	"net/http"
)

func handlePoll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	for _, s := range domain.GetSources() {
		err := p.Publish(ctx, "sources", SourceEvent{Source: s})
		if err != nil {
			httpError(ctx, w, "Error marshaling pubsub event", err)
			return
		}
		slog.Info(ctx, "Dispatched source %s", s.FeedURL)
	}
}
