package goose

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

type HtmlRequester interface {
	fetchHTML(string) (string, error)
}

// Crawler can fetch the target HTML page
type htmlrequester struct {
	config Configuration
	client http.Client
}

// NewCrawler returns a crawler object initialised with the URL and the [optional] raw HTML body
func NewHtmlRequester(config Configuration) HtmlRequester {
	jar, _ := cookiejar.New(nil)

	return htmlrequester{
		config: config,
		client: http.Client{
			Timeout: config.timeout,
			Jar:     jar, // binks
		},
	}
}

func (hr htmlrequester) fetchHTML(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/html")
	res, err := hr.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "could not perform request on "+url)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", &badRequest{Message: "could not perform request with " + url + " status code " + res.Status}
	}

	if !strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
		return "", errors.New("bad content type: " + res.Header.Get("Content-Type"))
	}

	buf, err := ioutil.ReadAll(res.Body)
	return string(buf), err
}

type badRequest struct {
	Message string `json:"message,omitempty"`
}

func (BadRequest *badRequest) Error() string {
	return BadRequest.Message
}
