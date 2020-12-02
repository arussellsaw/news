package handler

import (
	"net/http"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
)

func handleGenerateEdition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	e, err := dao.GetEditionForTime(ctx, time.Now(), false)
	if err != nil {
		slog.Error(ctx, "Error getting edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	if e != nil && r.URL.Query().Get("force") == "" {
		slog.Info(ctx, "Edition %s - %s already exists and is within window", e.ID, e.Name)
		w.Write([]byte(e.ID))
		return
	}

	e, err = domain.NewEdition(ctx, time.Now())
	if err != nil {
		slog.Error(ctx, "Error creating edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	start := time.Now().Add(-72 * time.Hour)
	end := time.Now()
	articles, err := dao.GetArticlesByTime(ctx, start, end)
	if err != nil {
		httpError(ctx, w, "error getting articles", err)
		return
	}

	newArticles := []domain.Article{}
L:
	for _, a := range articles {
		for _, e := range a.Content {
			if !utf8.Valid([]byte(e.Value)) {
				continue L
			}
		}
		newArticles = append(newArticles, a)
	}
	e.Articles = newArticles

	bySource := make(map[string][]domain.Article)
	for _, a := range e.Articles {
		bySource[a.Source.Name] = append(bySource[a.Source.Name], a)
	}
	newArticles = nil
	for _, as := range bySource {
		sort.Slice(as, func(i, j int) bool {
			return as[i].Timestamp.After(as[j].Timestamp)
		})
	}
top:
	for s, as := range bySource {
		newArticles = append(newArticles, as[0])
		bySource[s] = as[1:]
		if len(bySource[s]) == 0 {
			delete(bySource, s)
			goto top
		}
	}
	if len(bySource) != 0 {
		goto top
	}
	e.Articles = newArticles

	err = dao.SetEdition(ctx, e)
	if err != nil {
		slog.Error(ctx, "Error storing edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	slog.Info(ctx, "Created new edition: %s - %s", e.ID, e.Name)
	w.Write([]byte(e.ID))
}
