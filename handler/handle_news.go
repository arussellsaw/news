package handler

import (
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/monzo/slog"

	"github.com/arussellsaw/news/domain"
)

func handleNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	t := template.New("index.html")
	t, err := t.ParseFiles("tmpl/index.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	catMap := make(map[string]struct{})
	for _, s := range domain.GetSources() {
		for _, c := range s.Categories {
			catMap[c] = struct{}{}
		}
	}
	cats := []string{}
	for c := range catMap {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	articles, err := domain.GetArticles(ctx)
	if err != nil {
		slog.Error(ctx, "Error getting articles: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	articles = domain.FilterArticles(articles, r.URL.Query().Get("cat"))
	articles = domain.LayoutArticles(articles)

	sources := domain.GetSources()

	h := domain.Homepage{
		Title:      "The Webpage",
		Date:       time.Now().Format("Mon 02 Jan 2006"),
		Sources:    sources,
		Categories: cats,
		Articles:   articles,
	}
	err = t.Execute(w, h)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}
