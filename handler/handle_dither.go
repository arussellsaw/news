package handler

import (
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"

	dither "github.com/esimov/dithergo"
	"github.com/monzo/slog"
	"github.com/nfnt/resize"

	"github.com/arussellsaw/news/domain"
	"github.com/arussellsaw/news/pkg/util"
)

func handleDitherImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := r.URL.Query()
	url := q.Get("url")
	ctx = util.SetParam(ctx, "url", url)

	cachedImage := domain.C.GetImage(url)
	if cachedImage != nil {
		jpeg.Encode(w, *cachedImage, nil)
		return
	}

	res, err := http.Get(url)
	if err != nil {
		slog.Error(ctx, "Error getting image: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	width, err := strconv.ParseInt(q.Get("w"), 10, 64)
	if err != nil {
		slog.Error(ctx, "Error parsing width: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	img, _, err := image.Decode(res.Body)
	if err != nil {
		slog.Error(ctx, "Error decoding image: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	newImage := resize.Resize(uint(width), 0, img, resize.Lanczos3)

	d := dither.Dither{
		Type: "FloydSteinberg",
		Settings: dither.Settings{
			Filter: [][]float32{
				{0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0},
				{3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0},
				{1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0},
			},
		},
	}

	dithered := d.Monochrome(newImage, 1)

	w.Header().Set("Content-Type", "image/jpeg")

	err = jpeg.Encode(w, dithered, nil)
	if err != nil {
		slog.Error(ctx, "Error encoding image: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

}
