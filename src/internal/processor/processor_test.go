package processor

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestEncodeToTargetStaysWithinTargetByDownscaling(t *testing.T) {
	img := noisyImage(600, 600)

	_, smallestOriginal, err := encodeJPEGWithinTarget(img, 0)
	if err != nil {
		t.Fatalf("encodeJPEGWithinTarget: %v", err)
	}
	if smallestOriginal <= 4096 {
		t.Fatalf("expected original smallest jpeg to be larger than 4KB, got %d", smallestOriginal)
	}

	targetBytes := smallestOriginal / 3
	data, format, err := encodeToTarget(img, "jpeg", targetBytes)
	if err != nil {
		t.Fatalf("encodeToTarget: %v", err)
	}
	if format != "jpeg" {
		t.Fatalf("expected jpeg output, got %s", format)
	}
	if int64(len(data)) > targetBytes {
		t.Fatalf("expected output <= %d bytes, got %d", targetBytes, len(data))
	}

	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if decoded.Bounds().Dx() >= img.Bounds().Dx() && decoded.Bounds().Dy() >= img.Bounds().Dy() {
		t.Fatalf("expected downscaled output, got %dx%d from %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy(), img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestEncodeToTargetKeepsPNGWhenAlreadyWithinTarget(t *testing.T) {
	img := noisyImage(64, 64)

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	targetBytes := int64(pngBuf.Len() + 128)
	data, format, err := encodeToTarget(img, "png", targetBytes)
	if err != nil {
		t.Fatalf("encodeToTarget: %v", err)
	}
	if format != "png" {
		t.Fatalf("expected png output, got %s", format)
	}
	if int64(len(data)) > targetBytes {
		t.Fatalf("expected output <= %d bytes, got %d", targetBytes, len(data))
	}
}

func TestEncodeToTargetErrorsWhenTargetIsImpossible(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 30), B: 120, A: 255})
		}
	}

	if _, _, err := encodeToTarget(img, "jpeg", 1); err == nil {
		t.Fatal("expected error for impossible target, got nil")
	}
}

func noisyImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := uint8((x*37 + y*53 + (x*y)%97) % 256)
			img.Set(x, y, color.RGBA{
				R: value,
				G: uint8((int(value) + x*11) % 256),
				B: uint8((int(value) + y*17) % 256),
				A: 255,
			})
		}
	}
	return img
}

func TestEncodeJPEGWithinTargetReturnsHighestQualityThatFits(t *testing.T) {
	img := noisyImage(256, 256)

	var q80 bytes.Buffer
	if err := jpeg.Encode(&q80, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("jpeg encode q80: %v", err)
	}
	var q79 bytes.Buffer
	if err := jpeg.Encode(&q79, img, &jpeg.Options{Quality: 79}); err != nil {
		t.Fatalf("jpeg encode q79: %v", err)
	}

	targetBytes := int64(q80.Len())
	if int64(q79.Len()) > targetBytes {
		targetBytes = int64(q79.Len())
	}

	data, size, err := encodeJPEGWithinTarget(img, targetBytes)
	if err != nil {
		t.Fatalf("encodeJPEGWithinTarget: %v", err)
	}
	if data == nil {
		t.Fatal("expected a fitting jpeg result, got nil")
	}
	if size > targetBytes {
		t.Fatalf("expected size <= %d, got %d", targetBytes, size)
	}
}
