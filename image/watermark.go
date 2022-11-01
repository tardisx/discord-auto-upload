package image

import (
	"fmt"
	i "image"
	"image/jpeg"
	"image/png"
	"os"

	daulog "github.com/tardisx/discord-auto-upload/log"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/inconsolata"
)

// applyWatermark applies the watermark to the image
func (s *Store) applyWatermark() error {

	in, err := os.Open(s.uploadSourceFilename())

	defer in.Close()

	im, _, err := i.Decode(in)
	if err != nil {
		daulog.Errorf("Cannot decode image: %v - skipping watermarking", err)
		return fmt.Errorf("cannot decode image: %w", err)
	}
	bounds := im.Bounds()
	// var S float64 = float64(bounds.Max.X)

	dc := gg.NewContext(bounds.Max.X, bounds.Max.Y)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	dc.SetFontFace(inconsolata.Regular8x16)

	dc.DrawImage(im, 0, 0)

	dc.DrawRoundedRectangle(0, float64(bounds.Max.Y-18.0), 320, float64(bounds.Max.Y), 0)
	dc.SetRGB(0, 0, 0)
	dc.Fill()

	dc.SetRGB(1, 1, 1)

	dc.DrawString("github.com/tardisx/discord-auto-upload", 5.0, float64(bounds.Max.Y)-5.0)

	waterMarkedFile, err := os.CreateTemp("", "dau_watermark_file_*")

	if err != nil {
		return err
	}
	defer waterMarkedFile.Close()

	if s.OriginalFormat == "png" {
		png.Encode(waterMarkedFile, dc.Image())
	} else if s.OriginalFormat == "jpeg" {
		jpeg.Encode(waterMarkedFile, dc.Image(), nil)
	} else {
		panic("Cannot handle " + s.OriginalFormat)
	}

	s.WatermarkedFilename = waterMarkedFile.Name()
	return nil
}
