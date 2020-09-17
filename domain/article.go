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
