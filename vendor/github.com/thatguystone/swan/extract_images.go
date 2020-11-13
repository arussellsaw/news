package swan

import (
	"image"
	"math"
	"net/url"
	"strings"

	// So that we can read all the images
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html/atom"
)

type extractImages struct {
	a     *Article
	cache map[string]*Image
}

const (
	minImgWidth = 50
)

var (
	knownImgNames = []string{
		"yn-story-related-media",
		"cnn_strylccimg300cntr",
		"big_photo",
		"ap-smallphoto-a",
		"mediaimage",
	}
	badImgNames = []string{
		".html",
		".gif",
		".ico",
		"button",
		"twitter.jpg",
		"facebook.jpg",
		"ap_buy_photo",
		"digg.jpg",
		"digg.png",
		"delicious.png",
		"facebook.png",
		"reddit.jpg",
		"doubleclick",
		"diggthis",
		"diggThis",
		"adserver",
		"/ads/",
		"ec.atdmt.com",
		"mediaplex.com",
		"adsatt",
		"view.atdmt",
	}
	knownImgIds     = []cascadia.Selector{}
	knownImgClasses = []cascadia.Selector{}
	imgLink         = cascadia.MustCompile("link[rel=image_src][href]")
)

func init() {
	for _, n := range knownImgNames {
		knownImgIds = append(knownImgIds,
			cascadia.MustCompile("#"+n))
		knownImgClasses = append(knownImgClasses,
			cascadia.MustCompile("."+n))
	}
}

func hitImage(url string) *Image {
	i := Image{
		Src: url,
	}

	body, resp, err := httpGet(i.Src)
	if err != nil {
		return nil
	}

	defer body.Close()

	ri, _, err := image.Decode(body)
	if err != nil {
		return nil
	}

	i.Width = uint(ri.Bounds().Dx())
	i.Height = uint(ri.Bounds().Dy())
	if resp.ContentLength > 0 {
		i.Bytes = resp.ContentLength
	}

	return &i
}

func (e extractImages) run(a *Article) error {
	e.a = a
	e.cache = make(map[string]*Image)

	if e.checkKnown() {
		return nil
	}

	if a.TopNode == nil {
		return nil
	}

	if e.checkLarge(a.TopNode, 0) {
		return nil
	}

	if e.checkLinkTag() {
		return nil
	}

	e.checkOpenGraphTag()

	return nil
}

func (e *extractImages) hitCaches(imgs *goquery.Selection, attr string) []*Image {
	var hits []*Image

	imgs.Each(func(i int, img *goquery.Selection) {
		hit := e.hitCache(img, attr)
		if hit != nil {
			i := *hit
			i.Sel = img
			hits = append(hits, &i)
		}
	})

	return hits
}

func (e *extractImages) hitCacheURL(url string) *Image {
	i, ok := e.cache[url]
	if !ok {
		i = hitImage(url)
		e.cache[url] = i
	}

	return i
}

func (e *extractImages) hitCache(img *goquery.Selection, attr string) *Image {
	src, ok := img.Attr(attr)
	if !ok {
		return nil
	}

	url := e.buildURL(src)
	if url == "" {
		return nil
	}

	hit := e.hitCacheURL(url)
	if hit != nil {
		i := *hit
		i.Sel = img
		hit = &i
	}

	return hit
}

func (e *extractImages) getImage(img *goquery.Selection, attr string, c uint) *Image {
	i := e.hitCache(img, attr)
	if i != nil {
		i.Confidence = c
	}
	return i
}

func (e *extractImages) buildURL(src string) string {
	u, err := url.Parse(src)
	if err != nil {
		return ""
	}

	if u.IsAbs() {
		return src
	}

	return e.a.baseURL.ResolveReference(u).String()
}

func (e *extractImages) checkLarge(s *goquery.Selection, depth uint) bool {
	imgs := s.FindMatcher(imgTags).FilterFunction(
		func(i int, s *goquery.Selection) bool {
			if i > 30 {
				return false
			}

			src, ok := s.Attr("src")
			if !ok {
				return false
			}

			for _, s := range badImgNames {
				if strings.Contains(src, s) {
					return false
				}
			}

			return true
		}).FilterFunction(
		func(i int, s *goquery.Selection) bool {
			img := e.hitCache(s, "src")

			if img == nil {
				return false
			}

			return true
		})

	rimgs := e.hitCaches(imgs, "src")
	if len(rimgs) > 0 {
		var bestImg *Image

		cnt := 0
		initialArea := 0.0
		maxScore := 0.0

		if len(rimgs) > 30 {
			rimgs = rimgs[:30]
		}

		for _, i := range rimgs {
			shouldScore := ((depth >= 1 && i.Width > 300) || depth == 0) &&
				i.Width > minImgWidth &&
				!e.isBannerDims(i)
			if !shouldScore {
				continue
			}

			area := float64(i.Width * i.Height)
			score := 0.0

			if initialArea == 0.0 {
				initialArea = area * 1.48
				score = 1.0
			} else {
				areaDiff := area / initialArea
				sequenceScore := 1.0 / float64(cnt)
				score = sequenceScore * areaDiff
			}

			if score > maxScore {
				maxScore = score
				bestImg = i
			}

			cnt++
		}

		if bestImg != nil {
			bestImg.Confidence = uint(100 / len(rimgs))
			e.a.Img = bestImg
			return true
		}
	}

	if depth > 2 {
		return false
	}

	prev := s.Prev()
	if prev.Length() > 0 {
		return e.checkLarge(prev, depth)
	}

	par := s.Parent()
	if par.Length() > 0 {
		return e.checkLarge(par, depth+1)
	}

	return false
}

func (e *extractImages) isBannerDims(img *Image) bool {
	if img.Width == img.Height {
		return false
	}

	w := float64(img.Width)
	h := float64(img.Height)

	diff := math.Max(w, h) / math.Min(w, h)
	return diff > 5.0
}

func (e *extractImages) checkKnown() bool {
	check := func(s *goquery.Selection) (img *goquery.Selection) {
		s.EachWithBreak(func(i int, s *goquery.Selection) bool {
			if s.Nodes[0].DataAtom == atom.Img {
				img = s
				return false
			}

			c := s.FindMatcher(imgTags)
			if c.Size() > 0 {
				img = c.First()
			}

			return true
		})

		return
	}

	for _, cs := range knownImgIds {
		img := check(e.a.Doc.FindMatcher(cs))
		if img != nil {
			e.a.Img = e.getImage(img, "src", 90)
			return true
		}
	}

	for _, cs := range knownImgClasses {
		img := check(e.a.Doc.FindMatcher(cs))
		if img != nil {
			e.a.Img = e.getImage(img, "src", 90)
			return true
		}
	}

	return false
}

func (e *extractImages) checkLinkTag() bool {
	link := e.a.Doc.FindMatcher(imgLink)

	if link.Length() == 0 {
		return false
	}

	e.a.Img = e.getImage(link.First(), "href", 100)
	return true
}

func (e *extractImages) checkOpenGraphTag() bool {
	url, ok := e.a.Meta.OpenGraph["image"]
	if !ok {
		return false
	}

	i := e.hitCacheURL(url)
	if i != nil {
		i.Confidence = 100
		e.a.Img = i
	}

	return true
}
