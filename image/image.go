// Package image is responsible for thumbnailing, resizing and watermarking
// images.
package image

import (
	"fmt"
	i "image"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	daulog "github.com/tardisx/discord-auto-upload/log"

	"golang.org/x/image/draw"
)

// the filenames below are ordered in a specific way
// In the simplest case we only need the original filename.
// In more complex cases, we might have other files, created
// temporarily. These all need to be cleaned up.
// We upload the "final" file, depending on what actions have
// been taken.

type Store struct {
	OriginalFilename    string
	OriginalFormat      string // jpeg, png
	ModifiedFilename    string // if the user applied modifications
	ResizedFilename     string // if the file had to be resized to be uploaded
	WatermarkedFilename string
	MaxBytes            int
	Watermark           bool
}

// ReadCloser returns an io.ReadCloser providing the imagedata
// with the manglings that have been requested
func (s *Store) ReadCloser() (io.ReadCloser, error) {
	// determine format
	s.determineFormat()

	// check if we will fit the number of bytes, resize if necessary
	err := s.resizeToUnder(int64(s.MaxBytes))
	if err != nil {
		return nil, err
	}

	// the conundrum here is that the watermarking could modify the file size again, maybe going over
	//	the MaxBytes size. That would mostly be about jpeg compression levels I guess...
	if s.Watermark {
		s.applyWatermark()
	}

	// return the reader
	f, err := os.Open(s.uploadSourceFilename())
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (s *Store) determineFormat() error {
	file, err := os.Open(s.OriginalFilename)
	if err != nil {
		panic(fmt.Errorf("could not open file: %s", err))
	}
	defer file.Close()

	_, format, err := i.Decode(file)
	if err != nil {
		panic(fmt.Errorf("could not decode file: %s", err))
	}
	s.OriginalFormat = format
	return nil
}

// resizeToUnder resizes the image, if necessary
func (s *Store) resizeToUnder(size int64) error {
	fileToResize := s.uploadSourceFilename()
	fi, err := os.Stat(s.uploadSourceFilename())
	if err != nil {
		return err
	}
	currentSize := fi.Size()
	if currentSize <= size {
		return nil // nothing needs to be done
	}

	daulog.Infof("%s is %d bytes, need to resize to fit in %d", fileToResize, currentSize, size)

	file, err := os.Open(fileToResize)
	if err != nil {
		panic(fmt.Errorf("could not open file: %s", err))
	}
	defer file.Close()

	im, _, err := i.Decode(file)
	if err != nil {
		panic(fmt.Errorf("could not decode file: %s", err))
	}

	// if the size is 10% too big, we reduce X and Y by 10% - this is overkill but should
	// get us across the line in most cases
	fraction := float64(currentSize) / float64(size) // say 1.1 for 10%
	newXY := i.Point{
		X: int(float64(im.Bounds().Max.X) / fraction),
		Y: int(float64(im.Bounds().Max.Y) / fraction),
	}

	daulog.Infof("fraction is %f, will resize to %dx%d", fraction, newXY.X, newXY.Y)

	dst := i.NewRGBA(i.Rect(0, 0, newXY.X, newXY.Y))
	draw.BiLinear.Scale(dst, dst.Rect, im, im.Bounds(), draw.Over, nil)

	resizedFile, err := os.CreateTemp("", "dau_resize_file_*")
	if err != nil {
		return err
	}

	if s.OriginalFormat == "png" {
		err = png.Encode(resizedFile, dst)
		if err != nil {
			return err
		}
	} else if s.OriginalFormat == "jpeg" {
		err = jpeg.Encode(resizedFile, dst, nil)
		if err != nil {
			return err
		}

	} else {
		panic("unknown format " + s.OriginalFormat)
	}

	s.ResizedFilename = resizedFile.Name()
	resizedFile.Close()

	fi, err = os.Stat(s.uploadSourceFilename())
	if err != nil {
		return err
	}
	newSize := fi.Size()
	if newSize <= size {
		daulog.Infof("File resized, now %d", newSize)
		return nil // nothing needs to be done
	} else {
		return fmt.Errorf("failed to resize: was %d, now %d, needed %d", currentSize, newSize, size)
	}

}

// uploadSourceFilename gives us the filename, which might be a watermarked, resized
// or markedup version, depending on what has happened to this file.
func (s Store) uploadSourceFilename() string {
	if s.WatermarkedFilename != "" {
		return s.WatermarkedFilename
	}
	if s.ResizedFilename != "" {
		return s.ResizedFilename
	}
	if s.ModifiedFilename != "" {
		return s.ModifiedFilename
	}
	return s.OriginalFilename
}

// UploadFilename provides a name to be assigned to the upload on Discord
func (s Store) UploadFilename() string {
	return "image." + s.OriginalFormat
}

// Cleanup removes all the temporary files that we might have created
func (s Store) Cleanup() {
	daulog.Infof("cleaning temporary files %#v", s)

	if s.ModifiedFilename != "" {
		daulog.Infof("removing %s", s.ModifiedFilename)
		os.Remove(s.ModifiedFilename)
	}
	if s.ResizedFilename != "" {
		daulog.Infof("removing %s", s.ResizedFilename)
		os.Remove(s.ResizedFilename)
	}
	if s.WatermarkedFilename != "" {
		daulog.Infof("removing %s", s.WatermarkedFilename)
		os.Remove(s.WatermarkedFilename)
	}
}
