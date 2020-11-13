package swan

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	maxRespBytes = 15728640
)

var (
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
)

func httpGet(url string) (body io.ReadCloser, resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("could not create new request: %s", err)
		return
	}

	req.Header.Set("User-Agent", "swan/"+Version)
	resp, err = httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("could not load URL: %s", err)
		return
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		resp.Body = nil
		err = fmt.Errorf("could not load URL: status code %d", resp.StatusCode)
		return
	}

	body = http.MaxBytesReader(nil, resp.Body, maxRespBytes)
	return
}
