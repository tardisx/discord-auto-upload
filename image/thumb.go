package image

import (
	"fmt"
	i "image"
	"image/png"
	"io"
	"log"
	"os"

	"golang.org/x/image/draw"
)

const (
	thumbnailMaxX = 128
	thumbnailMaxY = 128
)

type ThumbType = string

const ThumbTypeOriginal = "orig"
const ThumbTypeMarkedUp = "markedup"

// ThumbPNG writes a thumbnail out to an io.Writer
func (ip *Store) ThumbPNG(t ThumbType, w io.Writer) error {

	var filename string
	if t == ThumbTypeOriginal {
		filename = ip.OriginalFilename
	} else if t == ThumbTypeMarkedUp {
		filename = ip.ModifiedFilename
	} else {
		log.Fatal("was passed incorrect 'type' arg")
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open file: %s", err)
	}
	defer file.Close()
	im, _, err := i.Decode(file)
	if err != nil {
		return fmt.Errorf("could not decode file: %s", err)
	}

	newXY := i.Point{}
	if im.Bounds().Max.X/thumbnailMaxX > im.Bounds().Max.Y/thumbnailMaxY {
		newXY.X = thumbnailMaxX
		newXY.Y = im.Bounds().Max.Y / (im.Bounds().Max.X / thumbnailMaxX)
	} else {
		newXY.Y = thumbnailMaxY
		newXY.X = im.Bounds().Max.X / (im.Bounds().Max.Y / thumbnailMaxY)
	}

	dst := i.NewRGBA(i.Rect(0, 0, newXY.X, newXY.Y))
	draw.BiLinear.Scale(dst, dst.Rect, im, im.Bounds(), draw.Over, nil)

	png.Encode(w, dst)

	return nil

}
