package handler

import (
	"github.com/arussellsaw/news/domain"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/arussellsaw/news/dao"
	"github.com/monzo/slog"
)

func handleArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	t := template.New("frame.html")
	t, err := t.ParseFiles("tmpl/frame.html", "tmpl/article.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	edition, err := dao.GetEditionForTime(ctx, time.Now(), true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	article, err := dao.GetArticle(ctx, strings.TrimPrefix(r.URL.Query().Get("id"), "art_"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	a := articlePage{
		Article:    article,
		Categories: edition.Categories,
		Name:       edition.Name,
		Date:       edition.Date,
	}
	err = t.Execute(w, a)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

type articlePage struct {
	*domain.Article
	Categories []string
	Name       string
	Date       string
}
