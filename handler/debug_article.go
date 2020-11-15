package handler

import (
	"encoding/json"
	"github.com/arussellsaw/news/dao"
	"github.com/monzo/slog"
	"github.com/thatguystone/swan"
	"io/ioutil"
	"net/http"
)

func handleDebugArticle(w http.ResponseWriter, r *http.Request) {
	var (
		ctx        = r.Context()
		slogParams = map[string]string{}
		url        = r.URL.Query().Get("url")
		id         = r.URL.Query().Get("id")
	)

	if id != "" {
		a, err := dao.GetArticle(ctx, id)
		if err != nil {
			slog.Error(ctx, "Error fetching article: %s", err, slogParams)
			return
		}
		url = a.Link
	}

	slog.Debug(ctx, "Debugging article %s", url)
	res, err := c.Get(url)
	if err != nil {
		slog.Error(ctx, "Error fetching article: %s", err, slogParams)
		return
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		slog.Error(ctx, "Error reading article: %s", err, slogParams)
		return
	}
	article, err := swan.FromHTML(url, buf)
	if err != nil {
		slog.Error(ctx, "Error parsing article: %s", err, slogParams)
		return
	}
	article.Doc = nil
	article.TopNode = nil
	out, err := json.MarshalIndent(article, "", "	")
	if err != nil {
		slog.Error(ctx, "Error marshaling response: %s", err, slogParams)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
