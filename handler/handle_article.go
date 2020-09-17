package handler

import (
	"html/template"
	"net/http"

	"github.com/arussellsaw/news/dao"
	"github.com/monzo/slog"
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

	article, err := dao.GetArticle(ctx, r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = t.Execute(w, article)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}
