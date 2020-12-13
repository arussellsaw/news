package handler

import (
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/monzo/slog"
	"html/template"
	"net/http"
)

type settingsPage struct {
	Sources []domain.Source
	base
}

type base struct {
	User       *domain.User
	Error      string
	Categories []string
	ID         string
	Name       string
	Title      string
	Meta       Meta
}

type Meta struct {
	Title       string
	Description string
	Image       string
	URL         string
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	t := template.New("frame.html")
	t, err := t.ParseFiles("tmpl/frame.html", "tmpl/meta.html", "tmpl/settings.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	u := domain.UserFromContext(ctx)
	if u == nil {
		http.Error(w, "Not logged in", 400)
		return
	}
	sources, err := dao.GetSources(ctx, u.ID)
	if err != nil {
		http.Error(w, "Couldn't get sources", 500)
		return
	}

	s := settingsPage{
		Sources: sources,
		base: base{
			ID:   "Settings",
			User: u,
		},
	}

	err = t.Execute(w, &s)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}
