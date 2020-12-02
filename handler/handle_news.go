package handler

import (
	"github.com/arussellsaw/news/dao"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/monzo/slog"

	"github.com/arussellsaw/news/domain"
)

type newsPage struct {
	base
	Articles []domain.Article

	claimed      map[string]bool
	cacheIndex   int
	DisableCache bool
}

func handleNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t := template.New("frame.html")
	t, err := t.ParseFiles("tmpl/frame.html", "tmpl/frontpage-1.html", "tmpl/section.html", "tmpl/article-tile.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	u := domain.UserFromContext(ctx)

	p := newsPage{
		base: base{
			User: u,
		},
	}
	if u != nil {
		articles, sources, err := dao.GetArticlesForOwner(ctx, u.ID, time.Now().Add(-48*time.Hour), time.Now())
		if err != nil {
			slog.Error(ctx, "Error getting edition: %s", err)
			http.Error(w, err.Error(), 500)
			return
		}
		byFeedURL := make(map[string]domain.Source)
		smap := make(map[string]struct{})
		for _, s := range sources {
			byFeedURL[s.FeedURL] = s
			for _, cat := range s.Categories {
				smap[cat] = struct{}{}
			}
		}
		for cat := range smap {
			p.Categories = append(p.Categories, cat)
		}
		sort.Strings(p.Categories)
		p.DisableCache = true
		newArticles := []domain.Article{}
	L:
		for _, a := range articles {
			for _, e := range a.Content {
				if !utf8.Valid([]byte(e.Value)) {
					continue L
				}
			}
			a.Source = byFeedURL[a.Source.FeedURL]
			newArticles = append(newArticles, a)
		}
		p.Articles = newArticles

		bySource := make(map[string][]domain.Article)
		for _, a := range p.Articles {
			bySource[a.Source.FeedURL] = append(bySource[a.Source.FeedURL], a)
		}
		newArticles = nil
		for _, as := range bySource {
			sort.Slice(as, func(i, j int) bool {
				return as[i].Timestamp.After(as[j].Timestamp)
			})
		}
		keys := []string{}
		for k := range bySource {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	top:
		for _, key := range keys {
			as, ok := bySource[key]
			if !ok {
				continue
			}
			newArticles = append(newArticles, as[0])
			bySource[key] = as[1:]
			if len(bySource[key]) == 0 {
				delete(bySource, key)
				goto top
			}
		}
		if len(bySource) != 0 {
			goto top
		}
		p.Articles = newArticles
	} else {
		e, err := dao.GetEditionForTime(ctx, time.Now(), true)
		if err != nil {
			slog.Error(ctx, "Error getting edition: %s", err)
			http.Error(w, err.Error(), 500)
			return
		}
		if e == nil {
			http.NotFound(w, r)
			return
		}
		p.Categories = e.Categories
		p.ID = e.ID
		p.Title = e.Date
		p.Name = e.Name
		p.Articles = e.Articles
	}

	cat := r.URL.Query().Get("cat")
	if cat != "" {
		sources, err := dao.GetSources(ctx, u.ID)
		if err != nil {
			http.Error(w, "Error getting sources", 500)
			return
		}
		sourceCats := make(map[string][]string)
		for _, s := range sources {
			sourceCats[s.FeedURL] = s.Categories
		}
		newArticles := []domain.Article{}
	articles:
		for _, a := range p.Articles {
			for _, c := range sourceCats[a.Source.FeedURL] {
				if c == cat {
					newArticles = append(newArticles, a)
					continue articles
				}
			}
		}
		p.Articles = newArticles
	}

	src := r.URL.Query().Get("src")
	if src != "" {
		newArticles := []domain.Article{}
		for _, a := range p.Articles {
			if a.Source.Name == src {
				newArticles = append(newArticles, a)
				continue
			}
		}
		p.Articles = newArticles
	}

	for i := range p.Articles {
		p.Articles[i].Description = removeHTMLTag(p.Articles[i].Description)
		p.Articles[i].Title = removeHTMLTag(p.Articles[i].Title)
	}

	err = t.Execute(w, &p)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func removeHTMLTag(in string) string {
	// regex to match html tag
	const pattern = `(<\/?[a-zA-A]+?[^>]*\/?>)*`
	r := regexp.MustCompile(pattern)
	groups := r.FindAllString(in, -1)
	// should replace long string first
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i]) > len(groups[j])
	})
	for _, group := range groups {
		if strings.TrimSpace(group) != "" {
			in = strings.ReplaceAll(in, group, "")
		}
	}
	return in
}

var lc = layoutCache{}

type layoutCache struct {
	mu    sync.RWMutex
	cache map[int]domain.Article
	id    string
}

func (c *layoutCache) Get(editionID string, i int) (domain.Article, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if editionID != c.id {
		c.cache = make(map[int]domain.Article)
	}
	a, ok := c.cache[i]
	return a, ok
}

func (c *layoutCache) Set(editionID string, i int, a domain.Article) {
	c.mu.Lock()
	if c.cache == nil {
		c.cache = make(map[int]domain.Article)
	}
	if editionID != c.id {
		c.cache = make(map[int]domain.Article)
	}
	c.cache[i] = a
	c.mu.Unlock()
}

func (e *newsPage) GetArticle(size int, image bool) domain.Article {
	if e.claimed == nil {
		e.claimed = make(map[string]bool)
	}
	if a, ok := lc.Get(e.ID, e.cacheIndex); ok && !e.DisableCache {
		e.cacheIndex++
		return a
	}
top:
	candidates := []domain.Article{}
	for _, a := range e.Articles {
		if a.Size() >= size {
			if a.ImageURL == "" && image {
				continue
			}
			if e.claimed[a.ID] {
				continue
			}
			candidates = append(candidates, a)
		}
	}
	if len(candidates) == 0 && size > 0 {
		size -= 100
		goto top
	} else if len(candidates) == 0 {
		return domain.Article{}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Timestamp.Before(candidates[i].Timestamp)
	})
	a := candidates[0]
	a.ImageURL = func() string {
		if image {
			return a.ImageURL
		}
		return ""
	}()

	e.claimed[a.ID] = true
	if !e.DisableCache {
		lc.Set(e.ID, e.cacheIndex, a)
	}
	e.cacheIndex++
	return a
}
