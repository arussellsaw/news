package handler

import (
	"fmt"
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/pacedotdev/firesearch-sdk/clients/go/firesearch"
	"net/http"
	"sort"
)

type result struct {
	Article *domain.Article
	HitText string
}

type searchPage struct {
	Results []result
	Query   string
	url     string
}

func handleSearch(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")

	searchResults, err := indexService.Search(ctx, firesearch.SearchRequest{
		Query: firesearch.SearchQuery{
			IndexPath: "news/search/articles",
			Text:      query,
			Limit:     200,
			Select:    []string{"title", "content"},
		},
	})
	if err != nil {
		return nil, err
	}
	p := searchPage{
		Query: query,
		url:   r.URL.String(),
	}
	for _, hit := range searchResults.Hits {
		article, err := dao.GetArticle(ctx, hit.ID)
		if err != nil {
			return nil, err
		}
		p.Results = append(p.Results, result{
			Article: article,
			HitText: hit.Highlights[0].Text,
		})
	}
	sort.Slice(p.Results, func(i, j int) bool {
		return p.Results[i].Article.Timestamp.After(p.Results[j].Article.Timestamp)
	})
	return p, nil
}

func (p searchPage) Meta() Meta {
	return Meta{
		Title: fmt.Sprintf("Search for %s on The Webpage", p.Query),
		URL:   p.url,
		Image: "/static/images/preview.png",
	}
}
