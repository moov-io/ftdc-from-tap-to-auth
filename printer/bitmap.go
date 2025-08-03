package printer

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
)

type BitmapImage struct {
	Width  int
	Height int
	Data   []byte // Bitmap data in 1-bit per pixel format
}

//go:embed logo.png
var logoData []byte

func NewLogoBitmap() (*BitmapImage, error) {
	img, err := png.Decode(bytes.NewReader(logoData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	bytesPerRow := (width + 7) / 8
	bitmapData := make([]byte, bytesPerRow*height)

	for y := range height {
		for x := range width {
			// Get pixel color
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to grayscale and determine if it's black
			isBlack := r < 32768 || g < 32768 || b < 32768

			if isBlack {
				// Set the bit for this pixel
				byteIndex := y*bytesPerRow + x/8
				bitIndex := 7 - (x % 8) // MSB first
				bitmapData[byteIndex] |= (1 << bitIndex)
			}
		}
	}

	// Load the bitmap image from file
	// This is a placeholder for actual image loading logic
	// You would typically use an image library to read the bitmap file
	return &BitmapImage{
		Width:  width,
		Height: height,
		Data:   bitmapData,
	}, nil
}
