package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/idgen"
	"github.com/monzo/slog"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/arussellsaw/news/domain"
)

type ArticleEvent struct {
	Article domain.Article
	Source  domain.Source
}

func handlePubsubArticle(w http.ResponseWriter, r *http.Request) {
	var (
		m   domain.PubSubMessage
		e   ArticleEvent
		ctx = r.Context()
	)
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		slog.Error(ctx, "Error decoding JSON: %s", err)
		return
	}

	if err := json.Unmarshal(m.Message.Data, &e); err != nil {
		slog.Error(ctx, "Error decoding JSON: %s", err)
		return
	}

	existing, err := dao.GetArticleByURL(ctx, e.Article.Link)
	if err != nil {
		slog.Error(ctx, "Error decoding JSON: %s", err)
		return
	}
	if existing != nil {
		slog.Debug(ctx, "Article already exists: %s - %s", existing.ID, existing.Title)
		return
	}

	slogParams := map[string]string{
		"url": e.Article.Link,
	}
	res, err := c.Get(fmt.Sprintf("https://readability-server.russellsaw.io/?url=%s", url.QueryEscape(e.Article.Link)))
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s", err, slogParams)
		return
	}
	var article = struct {
		Body     string `json:"body"`
		BodyText string `json:"body_text"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&article)
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s", err, slogParams)
		return
	}

	a := domain.Article{
		ID:          idgen.New("art"),
		Title:       e.Article.Title,
		Description: e.Article.Description,
		Content:     toElements(article.BodyText, "\n"),
		HTMLContent: template.HTML(article.Body),
		ImageURL:    e.Article.ImageURL,
		Link:        e.Article.Link,
		Source:      e.Article.Source,
		Timestamp:   e.Article.Timestamp,
		TS:          e.Article.Timestamp.Format("Mon Jan 2 15:04"),
	}
	err = dao.SetArticle(ctx, &a)
	if err != nil {
		slog.Error(ctx, "Error storing article: %s", err, slogParams)
		return
	}
	slog.Info(ctx, "Stored new article: %s - %s", a.ID, a.Title)
}

func httpError(ctx context.Context, w http.ResponseWriter, msg string, err error) {
	slog.Error(ctx, "%s: %s", msg, err)
	http.Error(w, err.Error(), 500)
	return
}

func toElements(s, br string) []domain.Element {
	var (
		lines = strings.Split(s, br)
		out   []domain.Element
		row   string
	)
	for _, line := range lines {
		if line == "p" {
			continue
		}
		if line == "" {
			out = append(out, domain.Element{Type: "text", Value: row})
			row = ""
			continue
		}
		row += line
	}
	if row != "" {
		out = append(out, domain.Element{Type: "text", Value: row})
	}
	return out
}
