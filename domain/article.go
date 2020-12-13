package domain

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/monzo/slog"
	"html/template"
	"io/ioutil"
	"time"
)

type Article struct {
	ID                string
	Title             string
	Description       string
	CompressedContent []byte
	Content           []Element
	ImageURL          string
	Link              string
	Author            string
	Source            Source
	Timestamp         time.Time
	TS                string
	Layout            Layout

	decompressed []byte
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

func (a *Article) RawHTML() template.HTML {
	if len(a.decompressed) != 0 {
		return template.HTML(a.decompressed)
	}
	slog.Debug(context.Background(), "Decompressing %s", a.ID)
	r, _ := gzip.NewReader(bytes.NewReader(a.CompressedContent))
	buf, _ := ioutil.ReadAll(r)
	a.decompressed = buf
	return template.HTML(buf)
}

func (a *Article) SetHTMLContent(body string) {
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	w.Write([]byte(body))
	w.Flush()
	w.Close()
	a.CompressedContent = buf.Bytes()
}
