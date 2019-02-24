package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"

	"golang.org/x/image/draw"
)

const dstWidth = 1024 // result file width
const dstHeight = 768 // result file height

func homeRouterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	// Read files
	fileImage, _, err := r.FormFile("image")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}
	defer fileImage.Close()

	body := &bytes.Buffer{}
	io.Copy(body, fileImage)

	fileWatermark, _, err := r.FormFile("watermark")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}
	defer fileWatermark.Close()

	bodyWatermark := &bytes.Buffer{}
	io.Copy(bodyWatermark, fileWatermark)

	// Make new image
	src, err := jpeg.Decode(body)
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	// Calculate data for image resizing
	// We need to determine rectangle resRect for draw.NearestNeighbor.Scale function
	sr := src.Bounds()
	srcWidth := sr.Max.X - sr.Min.X
	srcHeight := sr.Max.Y - sr.Min.Y
	// Calculate ratio for width and height
	var widthRatio, heightRatio float64
	widthRatio = float64(srcWidth) / dstWidth
	heightRatio = float64(srcHeight) / dstHeight
	// Get max ratio
	ratio := math.Max(widthRatio, heightRatio)
	resultWidth := int(math.Floor(float64(srcWidth) / ratio))
	resultHeight := int(math.Floor(float64(srcHeight) / ratio))
	var p image.Point
	p.X = int(math.Floor(float64(dstWidth-resultWidth) / 2.0))
	p.Y = int(math.Floor(float64(dstHeight-resultHeight) / 2.0))
	resRect := image.Rect(p.X, p.Y, p.X+resultWidth, p.Y+resultHeight)
	draw.NearestNeighbor.Scale(dst, resRect, src, src.Bounds(), draw.Over, nil)

	watermark, err := png.Decode(bodyWatermark)
	// Calculate tile numbers
	wmRect := watermark.Bounds()
	wmw := wmRect.Max.X - wmRect.Min.X
	wmh := wmRect.Max.Y - wmRect.Min.Y
	numw := (resultWidth/2 - wmw/2) / wmw
	restw := (resultWidth/2 - wmw/2) % wmw
	numh := (resultHeight/2 - wmh/2) / wmh
	resth := (resultHeight/2 - wmh/2) % wmh

	// Draw watermark tiles
	y := resth + p.Y
	for i := 0; i < numh*2+1; i++ {
		x := restw + p.X
		for j := 0; j < numw*2+1; j++ {
			draw.Copy(dst, image.Point{x, y}, watermark, watermark.Bounds(), draw.Over, nil)
			x += wmw
		}
		y += wmh
	}

	// write response image
	if jpeg.Encode(w, dst, nil) != nil {
		w.WriteHeader(500)
	}
}

func main() {
	http.HandleFunc("/watermark", homeRouterHandler)
	err := http.ListenAndServe(":3210", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
