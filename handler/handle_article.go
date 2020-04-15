package handler

import (
	"html/template"
	"net/http"

	"github.com/monzo/slog"

	"github.com/arussellsaw/news/domain"
)

func handleArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	t := template.New("article.html")
	t, err := t.ParseFiles("tmpl/article.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	article, ok := domain.GetArticleForLink(r.URL.Query().Get("url"))
	if !ok {
		http.NotFoundHandler().ServeHTTP(w, r)
	}

	err = t.Execute(w, article)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}
