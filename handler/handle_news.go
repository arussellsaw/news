package handler

import (
	"github.com/arussellsaw/news/dao"
	"html/template"
	"net/http"
	"time"

	"github.com/monzo/slog"

	"github.com/arussellsaw/news/domain"
)

func handleNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t := template.New("index.html")
	t, err := t.ParseFiles("tmpl/index.html", "tmpl/article-tile.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	e, err := dao.GetEditionForTime(ctx, time.Now(), true)
	if err != nil {
		slog.Error(ctx, "Error getting edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	if e == nil {
		http.NotFound(w, r)
		return
	}

	cat := r.URL.Query().Get("cat")
	if cat != "" {
		newArticles := []domain.Article{}
	articles:
		for _, a := range e.Articles {
			for _, c := range a.Source.Categories {
				if c == cat {
					newArticles = append(newArticles, a)
					continue articles
				}
			}
		}
		e.Articles = newArticles
	}

	src := r.URL.Query().Get("src")
	if src != "" {
		newArticles := []domain.Article{}
		for _, a := range e.Articles {
			if a.Source.Name == src {
				newArticles = append(newArticles, a)
				continue
			}
		}
		e.Articles = newArticles
	}

	err = t.Execute(w, &e)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}
