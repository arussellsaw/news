package domain

import "time"

type Article struct {
	ID          string
	Title       string
	Description string
	Content     []Element
	ImageURL    string
	Link        string
	Author      string
	Source      Source
	Timestamp   time.Time
	TS          string
	Layout      Layout
}

type Element struct {
	Type  string
	Value string
}

func (a *Article) Size() int {
	var n int
	for _, e := range a.Content {
		if e.Type == "text" {
			n += len(e.Value)
		}
	}
	return n
}

func (a *Article) Trim(size int) {
	oldE := a.Content
	a.Content = []Element{}
	var n int
	for _, e := range oldE {
		if e.Type == "text" {
			n += len(e.Value)
			if n > size {
				e.Value = string(e.Value[len(e.Value)-(n-size):]) + "..."
				return
			}
		}
		a.Content = append(a.Content, e)
	}
}
