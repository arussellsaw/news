package handler

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"

	dither "github.com/esimov/dithergo"
	"github.com/monzo/slog"
	"github.com/nfnt/resize"
)

func handleDitherImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := r.URL.Query()
	url := q.Get("url")
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

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, dithered, nil); err != nil {
		slog.Error(ctx, "Error encoding image: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		slog.Error(ctx, "Error writing image: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}
