package main

import (
	"bytes"
	dither "github.com/esimov/dithergo"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"strconv"
)

func handleDitherImage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	url := q.Get("url")
	res, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println("getting", err, url)
		return
	}

	width, err := strconv.ParseInt(q.Get("w"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err, url)
		return
	}

	img, _, err := image.Decode(res.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println("decode: ", err, url)
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
		log.Println("unable to encode image.", url)
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.", url)
	}
}
