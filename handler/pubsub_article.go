package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/idgen"
	"github.com/monzo/slog"
	"github.com/pacedotdev/firesearch-sdk/clients/go/firesearch"
	"github.com/thatguystone/swan"
	"golang.org/x/sync/semaphore"
	"net/http"
	"os/exec"
	"strings"

	"github.com/arussellsaw/news/domain"
)

var (
	s = semaphore.NewWeighted(20)
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
	err = s.Acquire(ctx, 1)
	if err != nil {
		slog.Error(ctx, "Failed to acquire semaphore: %s", err)
		return
	}
	cmd := exec.Command("node", "./readability-server/index.js", e.Article.Link)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	err = cmd.Run()
	s.Release(1)
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s - %s", err, buf.String(), slogParams)
		return
	}
	var article = struct {
		Body     string `json:"body"`
		BodyText string `json:"body_text"`
	}{}
	err = json.NewDecoder(buf).Decode(&article)
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s", err, slogParams)
		return
	}

	if len(article.Body) > 1024*1024*5 {
		slog.Warn(ctx, "dropping article, too large: %s", e.Article.Link)
		return
	}

	a := domain.Article{
		ID:          idgen.New("art"),
		Title:       removeHTMLTag(e.Article.Title),
		Description: removeHTMLTag(e.Article.Description),
		Content:     []domain.Element{{Type: "text", Value: removeHTMLTag(article.BodyText)}},
		ImageURL:    e.Article.ImageURL,
		Link:        e.Article.Link,
		Source:      e.Article.Source,
		Timestamp:   e.Article.Timestamp,
		TS:          e.Article.Timestamp.Format("Mon Jan 2 15:04"),
	}
	a.SetHTMLContent(article.Body)

	sa, err := swan.FromHTML(a.Link, []byte(article.Body))
	if err != nil {
		slog.Error(ctx, "Error parsing article: %s", err, slogParams)
		return
	}
	if sa.Img != nil {
		a.ImageURL = sa.Img.Src
	}

	err = dao.SetArticle(ctx, &a)
	if err != nil {
		slog.Error(ctx, "Error storing article: %s", err, slogParams)
		return
	}
	slog.Info(ctx, "Stored new article: %s - %s", a.ID, a.Title)
	_, err = indexService.PutDoc(ctx, firesearch.PutDocRequest{
		IndexPath: "news/search/articles",
		Doc: firesearch.Doc{
			ID: a.ID,
			SearchFields: []firesearch.SearchField{
				{
					Key:   "title",
					Value: a.Title,
					Store: true,
				},
				{
					Key:   "content",
					Value: article.BodyText,
					Store: true,
				},
			},
		},
	})
	if err != nil {
		slog.Error(ctx, "Error indexing: %s", err)
	}
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
