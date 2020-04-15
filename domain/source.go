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
		Name:       "TechCrunch",
		URL:        "https://techcrunch.com",
		FeedURL:    "http://feeds.feedburner.com/TechCrunch/",
		Categories: []string{"tech", "startups"},
	},
	{
		Name:       "BBC News",
		URL:        "https://bbc.co.uk/news",
		FeedURL:    "http://feeds.bbci.co.uk/news/rss.xml",
		Categories: []string{"news"},
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
		Categories: []string{"tech", "programming", "forums"},
	},
	{
		Name:       "Hackaday",
		URL:        "https://hackaday.com",
		FeedURL:    "https://hackaday.com/feed/",
		Categories: []string{"tech", "programming"},
	},
}
