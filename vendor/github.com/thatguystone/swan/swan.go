// Package swan implements the Goose HTML Content / Article Extractor
// algorithm.
//
// Currently, swan will try to extract the following content types:
//
// Comics: if something looks like a web comic, it will be extracted as just
// an image. This is a WIP.
//
// Everything else: it will look for article text and try to extract any
// header image that goes with it.
package swan

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/PuerkitoBio/goquery"
)

const (
	// Version of the library
	Version = "1.0"
)

// FromURL does its best to extract an article from the given URL
func FromURL(url string) (a *Article, err error) {
	body, resp, err := httpGet(url)
	if err != nil {
		return
	}

	defer body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("could not read response body: %s", err)
		return
	}

	return FromHTML(resp.Request.URL.String(), html)
}

// FromHTML does its best to extract an article from a single HTML page.
//
// Pass in the URL the document came from so that images can be resolved
// correctly.
func FromHTML(url string, html []byte) (*Article, error) {
	html, err := ToUtf8(html)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		err = fmt.Errorf("invalid HTML: %s", err)
		return nil, err
	}

	return FromDoc(url, doc)
}

// FromDoc does its best to extract an article from a single document
//
// Pass in the URL the document came from so that images can be resolved
// correctly.
func FromDoc(url string, doc *goquery.Document) (*Article, error) {
	a := &Article{
		URL: url,
		Doc: doc,
	}

	err := a.extract()
	if err != nil {
		return nil, err
	}

	return a, nil
}
