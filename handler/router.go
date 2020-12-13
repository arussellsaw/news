package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/cookiejar"
	"os"

	"cloud.google.com/go/profiler"
	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	secrets "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"

	"github.com/arussellsaw/news/domain"
	"github.com/arussellsaw/news/pkg/util"
	"github.com/felixge/fgprof"

	"github.com/gorilla/mux"
	"github.com/monzo/slog"
	"github.com/pacedotdev/firesearch-sdk/clients/go/firesearch"
)

var (
	jar, _ = cookiejar.New(&cookiejar.Options{})
	c      = http.Client{
		Jar: jar, //binks
	}

	p            domain.Publisher
	client       *firesearch.Client
	indexService *firesearch.IndexService
)

func Init(ctx context.Context) http.Handler {
	m := mux.NewRouter()
	m.Handle("/", http.HandlerFunc(handleNews))
	fs := http.FileServer(http.Dir("./static/"))
	m.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	m.Handle("/image", http.HandlerFunc(handleDitherImage))
	m.Handle("/article", http.HandlerFunc(handleArticle))
	m.Handle("/login", http.HandlerFunc(handleLogin))
	m.Handle("/settings", http.HandlerFunc(handleSettings))
	m.Handle("/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name: "sess",
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}))
	m.Handle("/favicon.ico", http.NotFoundHandler())
	m.Handle("/events/source", http.HandlerFunc(handlePubsubSource))
	m.Handle("/events/article", http.HandlerFunc(handlePubsubArticle))
	m.Handle("/poll", http.HandlerFunc(handlePoll))
	m.Handle("/article/debug", http.HandlerFunc(handleDebugArticle))
	m.Handle("/article/refresh", http.HandlerFunc(handleRefreshArticle))
	m.Handle("/settings/source", genericHandler("tmpl/settings_source.html", sourceSettingsData))
	m.Handle("/search", genericHandler("tmpl/search.html", handleSearch))
	m.Handle("/debug/fgprof", fgprof.Handler())
	cfg := profiler.Config{
		Service:        "news",
		ServiceVersion: "1.0.0",
		ProjectID:      "russellsaw",

		// For OpenCensus users:
		// To see Profiler agent spans in APM backend,
		// set EnableOCTelemetry to true
		// EnableOCTelemetry: true,
	}

	// Profiler initialization, best done as early as possible.
	if err := profiler.Start(cfg); err != nil {
		panic(err)
	}

	h := util.CloudContextMiddleware(
		util.HTTPLogParamsMiddleware(
			m,
		),
	)

	h = sessionMiddleware(h)

	var err error

	if os.Getenv("USER") == "alexrussell-saw" {
		slog.Info(ctx, "Using HTTP Publisher")
		p = &domain.HTTPPublisher{
			SourceURL:  "http://localhost:8080/events/source",
			ArticleURL: "http://localhost:8080/events/article",
		}
	} else {
		slog.Info(ctx, "Using PubSub Publisher")
		p, err = domain.NewPubSubPublisher(ctx)
		if err != nil {
			panic(err)
		}
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
	client = firesearch.NewClient("https://firesearch-3phpehgkya-ew.a.run.app/api", res.Payload.String())
	indexService = firesearch.NewIndexService(client)

	return h
}

type genericPage struct {
	Data interface{}
	base
}

func genericHandler(path string, data func(w http.ResponseWriter, r *http.Request) (interface{}, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u := domain.UserFromContext(ctx)
		p := genericPage{
			base: base{
				User: u,
			},
		}
		t := template.New("frame.html")
		t, err := t.ParseFiles("tmpl/frame.html", "tmpl/meta.html", path)
		if err != nil {
			slog.Error(ctx, "Error parsing template: %s", err)
			http.Error(w, err.Error(), 500)
			return
		}

		p.Data, err = data(w, r)
		if err != nil {
			p.Error = err.Error()
		}

		if mp, ok := p.Data.(MetaProvider); ok {
			p.Meta = mp.Meta()
		}

		err = t.Execute(w, &p)
		if err != nil {
			slog.Error(ctx, "Error executing template: %s", err)
			http.Error(w, err.Error(), 500)
			return
		}

	})
}

type MetaProvider interface {
	Meta() Meta
}
