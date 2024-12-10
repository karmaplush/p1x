package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
)

func loadImg(path string) (image.Image, error) {
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

func saveJpgImg(srcPath string, img image.Image, quality int) error {

	baseName := filepath.Base(srcPath)
	nameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]

	outputDir := "output"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.jpg", nameWithoutExt))

	file, err := os.Create(outputPath)
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
