package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/monzo/slog"

	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
)

func handleNews(w http.ResponseWriter, r *http.Request) {
	var (
		ctx        = r.Context()
		cat        = r.URL.Query().Get("cat")
		id         = r.URL.Query().Get("id")
		page       = r.URL.Query().Get("p")
		p    int64 = 1
		err  error
	)
	if page != "" {
		p, err = strconv.ParseInt(page, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	if id == "" {
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
		id := e.ID
		if cat != "" {
			http.Redirect(w, r, fmt.Sprintf("/?id=%s&cat=%s", id, cat), 302)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/?id=%s", id), 302)
		return
	}

	t := template.New("index.html")
	t, err = t.ParseFiles("tmpl/index.html")
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

	var e *domain.Edition

	e, err = dao.GetEdition(ctx, id)
	if err != nil {
		slog.Error(ctx, "Error getting edition: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	copyE := *e
	newArticles := []domain.Article{}
	for _, a := range copyE.Articles {
		for _, s := range a.Source.Categories {
			if cat == "" || s == cat {
				newArticles = append(newArticles, a)
				break
			}
		}
	}
	if cat != "" {
		newArticles = domain.LayoutArticles(newArticles)
	}

	e.Articles = paginate(newArticles, int(p))

	err = t.Execute(w, e)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func paginate(articles []domain.Article, p int) []domain.Article {
	out := []domain.Article{}
	row := 0
	rows := 0
	pageStart := (p - 1) * 5
	for _, a := range articles {
		row += a.Layout.Width
		if row > 12 {
			row = a.Layout.Width
			rows++
			if rows-pageStart == 5 {
				return out
			}
		}
		if rows >= pageStart {
			out = append(out, a)
		}
	}
	return out
}
