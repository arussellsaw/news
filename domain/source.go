package domain

import (
	"net/http"
	"net/http/cookiejar"
)

type Source struct {
	ID           string
	OwnerID      string
	Name         string
	URL          string
	FeedURL      string
	Categories   []string
	DisableFetch bool
}

var (
	jar, _ = cookiejar.New(&cookiejar.Options{})
	c      = http.Client{
		Jar: jar, //binks
	}
)

func GetSources() []Source {
	return sources
}

var sources = []Source{
	{
		Name:       "Vox",
		URL:        "https://vox.com",
		FeedURL:    "https://www.vox.com/rss/index.xml",
		Categories: []string{"news", "opinion"},
	},
	{
		Name:       "New York Times",
		URL:        "https://nytimes.com",
		FeedURL:    "https://rss.nytimes.com/services/xml/rss/nyt/HomePage.xml",
		Categories: []string{"news", "opinion"},
	},
	{
		Name:       "The Verge",
		URL:        "https://theverge.com",
		FeedURL:    "https://www.theverge.com/rss/index.xml",
		Categories: []string{"tech", "electronics"},
	},
	{
		Name:       "Polygon",
		URL:        "https://polygon.com",
		FeedURL:    "https://www.polygon.com/rss/index.xml",
		Categories: []string{"games"},
	},
	{
		Name:       "The Guardian",
		URL:        "https://theguardian.co.uk",
		FeedURL:    "https://www.theguardian.com/uk/rss",
		Categories: []string{"news", "opinion"},
	},
	{
		Name:       "lobste.rs",
		URL:        "https://lobste.rs",
		FeedURL:    "https://lobste.rs/rss",
		Categories: []string{"programming", "forums"},
	},
	{
		Name:       "Hackaday",
		URL:        "https://hackaday.com",
		FeedURL:    "https://hackaday.com/feed/",
		Categories: []string{"programming"},
	},
	{
		Name:       "Hacker News",
		URL:        "https://news.ycombinator.com",
		FeedURL:    "https://news.ycombinator.com/rss",
		Categories: []string{"programming", "forums"},
	},
	{
		Name:       "The Atlantic",
		URL:        "https://theatlantic.com",
		FeedURL:    "https://www.theatlantic.com/feed/all/",
		Categories: []string{"opinion", "news"},
	},
	{
		Name:         "Huffington Post",
		URL:          "https://huffpost.com",
		FeedURL:      "https://www.huffpost.com/section/front-page/feed",
		Categories:   []string{"opinion", "news"},
		DisableFetch: true,
	},
	{
		Name:       "Eurogamer",
		URL:        "https://eurogamer.net",
		FeedURL:    "https://www.eurogamer.net/?format=rss",
		Categories: []string{"games"},
	},
	{
		Name:       "Bon Appetit",
		URL:        "https://bonappetit.com",
		FeedURL:    "https://www.bonappetit.com/feed/rss",
		Categories: []string{"food"},
	},
	{
		Name:       "Food52",
		URL:        "https://food52.com",
		FeedURL:    "https://food52.com/blog.rss",
		Categories: []string{"food"},
	},
	{
		Name:       "AnandTech",
		URL:        "https://anandtech.com",
		FeedURL:    "https://www.anandtech.com/rss/",
		Categories: []string{"tech", "electronics"},
	},
	{
		Name:       "The Register",
		URL:        "https://theregister.com",
		FeedURL:    "https://www.theregister.com/headlines.atom",
		Categories: []string{"tech", "programming"},
	},
	{
		Name:       "Ars Technica",
		URL:        "https://arstechnica.com",
		FeedURL:    "http://feeds.arstechnica.com/arstechnica/index",
		Categories: []string{"tech", "programming"},
	},
}
