package domain

type Source struct {
	Name         string
	URL          string
	FeedURL      string
	Categories   []string
	ForceFetch   bool
	DisableFetch bool
}

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
		Categories: []string{"tech", "games", "electronics"},
		ForceFetch: false,
	},
	{
		Name:       "Polygon",
		URL:        "https://polygon.com",
		FeedURL:    "https://www.polygon.com/rss/index.xml",
		Categories: []string{"tech", "games"},
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
}
