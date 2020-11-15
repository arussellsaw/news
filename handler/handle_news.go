package handler

import (
	"github.com/arussellsaw/news/dao"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/monzo/slog"

	"github.com/arussellsaw/news/domain"
)

func handleNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t := template.New("frame.html")
	t, err := t.ParseFiles("tmpl/frame.html", "tmpl/frontpage-1.html", "tmpl/section.html", "tmpl/article-tile.html")
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

	for i := range e.Articles {
		e.Articles[i].Description = removeHTMLTag(e.Articles[i].Description)
	}

	err = t.Execute(w, &e)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func removeHTMLTag(in string) string {
	// regex to match html tag
	const pattern = `(<\/?[a-zA-A]+?[^>]*\/?>)*`
	r := regexp.MustCompile(pattern)
	groups := r.FindAllString(in, -1)
	// should replace long string first
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i]) > len(groups[j])
	})
	for _, group := range groups {
		if strings.TrimSpace(group) != "" {
			in = strings.ReplaceAll(in, group, "")
		}
	}
	return in
}
