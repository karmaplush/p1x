package io

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"time"

	"github.com/karmaplush/p1x/converter"
)

func OpenImage(path string) (image.Image, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file opening error: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("file decoding error: %w", err)
	}

	return img, nil
}

func MustOpenImageAsNRGBA(path string) *image.NRGBA {

	srcImg, err := OpenImage(path)
	if err != nil {
		panic(err)
	}

	img, err := converter.ConvertToNRGBA(srcImg)
	if err != nil {
		panic(err)
	}

	return img
}

func SaveImageAsJPG(img *image.NRGBA, dir string, quality int) error {

	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf(
		"%s/p1x_%d.jpg",
		dir,
		timestamp,
	)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("file creation error: %w", err)
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return fmt.Errorf("image encoding error: %w", err)
	}

	return nil

}
