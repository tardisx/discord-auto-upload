package imageprocess

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/tardisx/discord-auto-upload/upload"
	"golang.org/x/image/draw"
)

const (
	thumbnailMaxX = 128
	thumbnailMaxY = 128
)

func (ip *Processor) ThumbPNG(ul *upload.Upload, which string, w io.Writer) error {

	var filename string
	if which == "orig" {
		filename = ul.OriginalFilename
	} else if which == "markedup" {
		filename = ul.MarkedUpFilename
	} else {
		log.Fatal("was passed incorrect 'which' arg")
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open file: %s", err)
	}
	defer file.Close()
	im, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("could not decode file: %s", err)
	}

	newXY := image.Point{}
	if im.Bounds().Max.X/thumbnailMaxX > im.Bounds().Max.Y/thumbnailMaxY {
		newXY.X = thumbnailMaxX
		newXY.Y = im.Bounds().Max.Y / (im.Bounds().Max.X / thumbnailMaxX)
	} else {
		newXY.Y = thumbnailMaxY
		newXY.X = im.Bounds().Max.X / (im.Bounds().Max.Y / thumbnailMaxY)
	}

	dst := image.NewRGBA(image.Rect(0, 0, newXY.X, newXY.Y))
	draw.BiLinear.Scale(dst, dst.Rect, im, im.Bounds(), draw.Over, nil)

	png.Encode(w, dst)

	return nil

}
