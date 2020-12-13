package main

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	"context"
	"flag"
	"fmt"
	"github.com/arussellsaw/news/pkg/util"
	"github.com/monzo/slog"
	"github.com/pacedotdev/firesearch-sdk/clients/go/firesearch"
	secrets "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	"os"
)

func main() {
	ctx := context.Background()

	var logger slog.Logger
	logger = util.ContextParamLogger{Logger: &util.StackDriverLogger{}}
	logger = util.ColourLogger{Writer: os.Stdout}
	slog.SetDefaultLogger(logger)

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

	query := flag.String("q", "", "query")
	flag.Parse()
	searchResults, err := indexService.Search(ctx, firesearch.SearchRequest{
		Query: firesearch.SearchQuery{
			IndexPath: "news/search/articles",
			Limit:     5,
			Text:      *query,
			Select:    []string{"title", "content"},
		},
	})
	if err != nil {
		slog.Critical(ctx, "Error searching: %s", err)
		return
	}
	for _, hit := range searchResults.Hits {
		title, ok := hit.FieldValue("title")
		if !ok {
			title = "Untitled"
		}
		fmt.Printf("\t%s: %s:", hit.ID, title)
		for _, highlight := range hit.Highlights {
			fmt.Print(" " + highlight.Text)
		}
		fmt.Println()
	}
}
