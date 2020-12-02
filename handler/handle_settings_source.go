package handler

import (
	"fmt"
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/arussellsaw/news/idgen"
	"net/http"
	"strings"
)

type sourceSettingsPage struct {
	domain.Source
	Action           string
	CategoriesString string
}

func sourceSettingsData(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	var (
		id         = r.Form.Get("id")
		name       = r.Form.Get("name")
		action     = r.Form.Get("action")
		confirm    = r.Form.Get("confirm")
		categories = r.Form.Get("categories")
		url        = r.Form.Get("url")
		feedURL    = r.Form.Get("feed_url")

		source *domain.Source
		u      = domain.UserFromContext(ctx)
		err    error
		cats   string
	)

	if id != "" {
		source, err = dao.GetSource(ctx, id)
		if err != nil {
			return nil, err
		}
		if source.OwnerID != u.ID {
			return nil, fmt.Errorf("permission denied")
		}
		cats = strings.Join(source.Categories, ",")
	}
	if action == "delete" && confirm == "true" {
		err = dao.DeleteSource(ctx, id)
		if err != nil {
			return nil, err
		}
		http.Redirect(w, r, "/settings", http.StatusTemporaryRedirect)
	}
	if action == "edit" && confirm == "true" {
		if id == "" {
			id = idgen.New("src")
		}
		categories := strings.Split(categories, ",")
		for i := range categories {
			categories[i] = strings.TrimSpace(categories[i])
		}
		src := domain.Source{
			Name:       name,
			ID:         id,
			OwnerID:    u.ID,
			URL:        url,
			FeedURL:    feedURL,
			Categories: categories,
		}
		err = dao.SetSource(ctx, &src)
		if err != nil {
			return nil, err
		}
		source = &src
		http.Redirect(w, r, "/settings", http.StatusTemporaryRedirect)
	}
	return sourceSettingsPage{
		Source: func() domain.Source {
			if source != nil {
				return *source
			}
			return domain.Source{}
		}(),
		CategoriesString: cats,
		Action:           action,
	}, nil
}
