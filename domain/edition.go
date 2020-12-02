package domain

import (
	"context"
	"github.com/arussellsaw/news/idgen"
	"sync"
	"time"
)

var (
	morningEdition time.Duration = 6 * time.Hour
	eveningEdition time.Duration = 17 * time.Hour

	lc = layoutCache{}
)

type layoutCache struct {
	mu    sync.RWMutex
	cache map[int]Article
	id    string
}

func (c *layoutCache) Get(editionID string, i int) (Article, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if editionID != c.id {
		c.cache = make(map[int]Article)
	}
	a, ok := c.cache[i]
	return a, ok
}

func (c *layoutCache) Set(editionID string, i int, a Article) {
	c.mu.Lock()
	if c.cache == nil {
		c.cache = make(map[int]Article)
	}
	if editionID != c.id {
		c.cache = make(map[int]Article)
	}
	c.cache[i] = a
	c.mu.Unlock()
}

type Edition struct {
	ID         string
	Name       string
	Sources    []Source
	Articles   []Article
	Categories []string
	Date       string

	StartTime time.Time
	EndTime   time.Time
	Created   time.Time

	Metadata map[string]string

	Article      Article
	claimed      map[string]bool
	cacheIndex   int
	DisableCache bool
}

func NewEdition(ctx context.Context, now time.Time) (*Edition, error) {
	morning := now.Truncate(24 * time.Hour).Add(morningEdition)
	evening := now.Truncate(24 * time.Hour).Add(eveningEdition)

	catMap := make(map[string]struct{})
	for _, s := range sources {
		for _, c := range s.Categories {
			catMap[c] = struct{}{}
		}
	}
	cats := make([]string, 0, len(catMap))
	for c := range catMap {
		cats = append(cats, c)
	}

	e := Edition{
		ID:         idgen.New("edt"),
		Sources:    sources,
		Categories: cats,
		Created:    time.Now(),
	}

	e.Date = time.Now().Format("Monday January 02 2006")

	switch {
	case now.After(morning) && now.Before(evening):
		e.StartTime = morning
		e.EndTime = evening
		e.Name = "Morning Edition"
	case now.Before(morning) || now.After(evening):
		e.StartTime = evening
		e.EndTime = morning.Add(24 * time.Hour)
		e.Name = "Evening Edition"
	}

	return &e, nil
}
