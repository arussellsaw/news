package handler

import (
	"context"
	"encoding/json"
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/idgen"
	"github.com/thatguystone/swan"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/monzo/slog"

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
	var content string
	res, err := c.Get(e.Article.Link)
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s", err, slogParams)
		return
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		slog.Error(ctx, "Error reading article: %s", err, slogParams)
		return
	}
	article, err := swan.FromHTML(e.Article.Link, buf)
	if err != nil {
		slog.Error(ctx, "Error parsing article: %s", err, slogParams)
		return
	}
	content = article.CleanedText
	if !utf8.Valid([]byte(content)) {
		slog.Error(ctx, "Skipping invalid utf8 in document: %s - %s", e.Article.Link, e.Article.Title)
		return
	}

	a := domain.Article{
		ID:          idgen.New("art"),
		Title:       e.Article.Title,
		Description: article.Meta.Description,
		Content:     toElements(content, "\n"),
		ImageURL: func() string {
			if article.Img != nil {
				return article.Img.Src
			}
			return ""
		}(),
		Link:      e.Article.Link,
		Source:    e.Article.Source,
		Timestamp: time.Now(),
		TS:        time.Now().Format("Mon Jan 2 15:04"),
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
