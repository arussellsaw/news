package handler

import (
	"fmt"
	"image/jpeg"
	"net/http"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/monzo/slog"
)

func handleQRCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.URL.Query().Get("id")

	b, err := qr.Encode(fmt.Sprintf("https://news.russellsaw.io/?id=%s", id), qr.M, qr.Auto)

	b, _ = barcode.Scale(b, 100, 100)

	if err != nil {
		slog.Error(ctx, "Error encoding barcode: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}

	err = jpeg.Encode(w, b, nil)
	if err != nil {
		slog.Error(ctx, "Error writing image: %s", err)
	}
}
