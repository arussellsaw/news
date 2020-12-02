package handler

import (
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
	"html/template"
	"net/http"
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

	article, err := dao.GetArticle(ctx, r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	u := domain.UserFromContext(ctx)
	if u != nil {
		sources, err := dao.GetSources(ctx, u.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		for _, src := range sources {
			if src.FeedURL == article.Source.FeedURL {
				article.Source = src
			}
		}
	}

	a := articlePage{
		Article: article,
		base: base{
			User: domain.UserFromContext(ctx),
		},
	}
	err = t.Execute(w, a)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

type articlePage struct {
	Article *domain.Article
	base
}
