package main

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	"context"
	"fmt"
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/arussellsaw/news/pkg/util"
	"github.com/monzo/slog"
	"github.com/pacedotdev/firesearch-sdk/clients/go/firesearch"
	secrets "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	"os"
	"time"
)

func main() {
	ctx := context.Background()

	var logger slog.Logger
	logger = util.ContextParamLogger{Logger: &util.StackDriverLogger{}}
	logger = util.ColourLogger{Writer: os.Stdout}
	slog.SetDefaultLogger(logger)

	err := dao.Init(ctx)
	if err != nil {
		slog.Critical(ctx, "Error setting up dao: %s", err)
		return
	}

	sm, err := secretmanager.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	defer sm.Close()

	res, err := sm.AccessSecretVersion(
		ctx,
		&secrets.AccessSecretVersionRequest{Name: fmt.Sprintf(
			"projects/266969078315/secrets/%s/versions/latest",
			"FIRESEARCH_API_KEY",
		)},
	)
	if err != nil {
		panic(err)
	}

	slog.Info(ctx, res.Payload.String())
	client := firesearch.NewClient("https://firesearch-3phpehgkya-ew.a.run.app/api", res.Payload.String())
	indexService := firesearch.NewIndexService(client)

	c := dao.Client()
	docs := c.Collection("articles").Documents(ctx)
	n := 0
	for {
		n++
		start := time.Now()
		doc, err := docs.Next()
		if err != nil {
			slog.Critical(ctx, "error reading article: %s", err)
			slog.Info(ctx, "done %v articles", n)
			return
		}
		if doc == nil {
			slog.Info(ctx, "done %v articles", n)
			return
		}
		a := domain.Article{}
		err = doc.DataTo(&a)
		if err != nil {
			slog.Critical(ctx, "error reading article: %s", err)
			return
		}
		a.RawHTML()
		contentStr := ""
		for _, e := range a.Content {
			if e.Type != "text" {
				continue
			}
			contentStr = contentStr + e.Value + ""
		}
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
						Value: contentStr,
						Store: true,
					},
				},
			},
		})
		if err != nil {
			slog.Error(ctx, "Error indexing: %s", err)
		}
		slog.Info(ctx, "Article: %s %s", time.Since(start), a.Link)
	}
}
