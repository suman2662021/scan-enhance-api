package processor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"strings"

	"image-compressar/src/internal/domain"
)

type Result struct {
	Data       []byte
	Format     string
	Width      int
	Height     int
	SHA256     string
	FinalBytes int64
}

type Processor interface {
	Process(input []byte, mimeType string, req domain.ProcessRequest) (Result, error)
}

type DocumentProcessor struct{}

func NewDocumentProcessor() *DocumentProcessor {
	return &DocumentProcessor{}
}

func (p *DocumentProcessor) Process(input []byte, mimeType string, req domain.ProcessRequest) (Result, error) {
	img, format, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return Result{}, fmt.Errorf("decode image: %w", err)
	}

	rgba := toRGBA(img)
	enhanced := p.applyEnhancements(rgba, req.Enhancement)
	outputFormat := normalizeFormat(req.OutputFormat, format)

	targetBytes := int64(req.TargetSizeKB) * 1024
	if targetBytes <= 0 {
		targetBytes = int64(len(input))
	}

	data, finalFormat, err := encodeToTarget(enhanced, outputFormat, targetBytes)
	if err != nil {
		return Result{}, err
	}

	hash := sha256.Sum256(data)
	bounds := enhanced.Bounds()

	return Result{
		Data:       data,
		Format:     finalFormat,
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
		SHA256:     hex.EncodeToString(hash[:]),
		FinalBytes: int64(len(data)),
	}, nil
}

func (p *DocumentProcessor) applyEnhancements(img *image.RGBA, opts domain.EnhancementOptions) *image.RGBA {
	current := img

	if opts.BrightnessAdjust || opts.Mode == "document_scan" {
		current = adjustBrightness(current, 10)
	}
	if opts.ContrastBoost || opts.Mode == "document_scan" {
		current = adjustContrast(current, 1.15)
	}
	if opts.ShadowRemoval || opts.Mode == "document_scan" {
		current = removeShadows(current)
	}
	if opts.Denoise || opts.Mode == "document_scan" {
		current = boxBlur(current, 1)
	}
	if opts.Sharpen || opts.Mode == "document_scan" {
		current = sharpen(current)
	}
	if opts.Grayscale || opts.BlackAndWhite || opts.Mode == "document_scan" {
		current = grayscale(current)
	}
	if opts.BlackAndWhite {
		current = threshold(current, 160)
	}

	return current
}

func normalizeFormat(requested, decoded string) string {
	switch strings.ToLower(strings.TrimSpace(requested)) {
	case "jpg", "jpeg":
		return "jpeg"
	case "png":
		return "png"
	default:
		switch strings.ToLower(decoded) {
		case "jpeg", "jpg", "png":
			return strings.ToLower(decoded)
		default:
			return "jpeg"
		}
	}
}

func encodeToTarget(img image.Image, format string, targetBytes int64) ([]byte, string, error) {
	if format == "png" {
		var buf bytes.Buffer
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		if int64(buf.Len()) <= targetBytes {
			return buf.Bytes(), "png", nil
		}
		format = "jpeg"
	}

	working := img
	for attempt := 0; attempt < 8; attempt++ {
		best, smallestSize, err := encodeJPEGWithinTarget(working, targetBytes)
		if err != nil {
			return nil, "", err
		}
		if best != nil {
			return best, "jpeg", nil
		}

		next := downscaleForTarget(working, targetBytes, smallestSize)
		if next.Bounds().Dx() == working.Bounds().Dx() && next.Bounds().Dy() == working.Bounds().Dy() {
			break
		}
		working = next
	}

	return nil, "", fmt.Errorf("unable to encode image within %d bytes", targetBytes)
}

func encodeJPEGWithinTarget(img image.Image, targetBytes int64) ([]byte, int64, error) {
	var smallest []byte
	smallestSize := int64(-1)

	for quality := 100; quality >= 1; quality-- {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, 0, err
		}
		size := int64(buf.Len())
		if size <= targetBytes {
			return append([]byte(nil), buf.Bytes()...), size, nil
		}
		if smallestSize == -1 || size < smallestSize {
			smallestSize = size
			smallest = append([]byte(nil), buf.Bytes()...)
		}
	}

	if smallest == nil {
		return nil, 0, errors.New("failed to encode output image")
	}
	return nil, smallestSize, nil
}

func downscaleForTarget(img image.Image, targetBytes, smallestSize int64) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 1 && height <= 1 {
		return img
	}

	scaleNumerator := targetBytes * 100
	scaleDenominator := smallestSize
	if scaleDenominator <= 0 {
		scaleDenominator = targetBytes
	}

	scalePercent := int(integerSqrt((scaleNumerator * 100) / scaleDenominator))
	if scalePercent > 95 {
		scalePercent = 95
	}
	if scalePercent < 50 {
		scalePercent = 50
	}

	newWidth := width * scalePercent / 100
	newHeight := height * scalePercent / 100
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}
	if newWidth == width && width > 1 {
		newWidth = width - 1
	}
	if newHeight == height && height > 1 {
		newHeight = height - 1
	}

	return resizeNearest(img, newWidth, newHeight)
}

func resizeNearest(src image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	for y := 0; y < height; y++ {
		srcY := srcBounds.Min.Y + (y*srcHeight)/height
		if srcY >= srcBounds.Max.Y {
			srcY = srcBounds.Max.Y - 1
		}
		for x := 0; x < width; x++ {
			srcX := srcBounds.Min.X + (x*srcWidth)/width
			if srcX >= srcBounds.Max.X {
				srcX = srcBounds.Max.X - 1
			}
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}

func integerSqrt(n int64) int64 {
	if n <= 0 {
		return 0
	}

	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	return x
}

func toRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}

func adjustBrightness(src *image.RGBA, delta int) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			dst.Set(x, y, color.RGBA{
				R: clamp8(int(r>>8) + delta),
				G: clamp8(int(g>>8) + delta),
				B: clamp8(int(b>>8) + delta),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func adjustContrast(src *image.RGBA, factor float64) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			dst.Set(x, y, color.RGBA{
				R: contrastChannel(uint8(r>>8), factor),
				G: contrastChannel(uint8(g>>8), factor),
				B: contrastChannel(uint8(b>>8), factor),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func removeShadows(src *image.RGBA) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var total int
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			total += int((r>>8)+(g>>8)+(b>>8)) / 3
		}
		rowAvg := total / bounds.Dx()
		offset := 220 - rowAvg
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			dst.Set(x, y, color.RGBA{
				R: clamp8(int(r>>8) + offset),
				G: clamp8(int(g>>8) + offset),
				B: clamp8(int(b>>8) + offset),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func grayscale(src *image.RGBA) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			luma := uint8((299*(r>>8) + 587*(g>>8) + 114*(b>>8)) / 1000)
			dst.Set(x, y, color.RGBA{R: luma, G: luma, B: luma, A: uint8(a >> 8)})
		}
	}
	return dst
}

func threshold(src *image.RGBA, cutoff uint8) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, a := src.At(x, y).RGBA()
			value := uint8(0)
			if uint8(r>>8) >= cutoff {
				value = 255
			}
			dst.Set(x, y, color.RGBA{R: value, G: value, B: value, A: uint8(a >> 8)})
		}
	}
	return dst
}

func boxBlur(src *image.RGBA, radius int) *image.RGBA {
	if radius <= 0 {
		return src
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rs, gs, bs, as, count int
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx := x + dx
					ny := y + dy
					if nx < bounds.Min.X || nx >= bounds.Max.X || ny < bounds.Min.Y || ny >= bounds.Max.Y {
						continue
					}
					r, g, b, a := src.At(nx, ny).RGBA()
					rs += int(r >> 8)
					gs += int(g >> 8)
					bs += int(b >> 8)
					as += int(a >> 8)
					count++
				}
			}
			dst.Set(x, y, color.RGBA{
				R: uint8(rs / count),
				G: uint8(gs / count),
				B: uint8(bs / count),
				A: uint8(as / count),
			})
		}
	}
	return dst
}

func sharpen(src *image.RGBA) *image.RGBA {
	blurred := boxBlur(src, 1)
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			or, og, ob, oa := src.At(x, y).RGBA()
			br, bg, bb, _ := blurred.At(x, y).RGBA()
			dst.Set(x, y, color.RGBA{
				R: clamp8(int(or>>8) + (int(or>>8) - int(br>>8))),
				G: clamp8(int(og>>8) + (int(og>>8) - int(bg>>8))),
				B: clamp8(int(ob>>8) + (int(ob>>8) - int(bb>>8))),
				A: uint8(oa >> 8),
			})
		}
	}
	return dst
}

func contrastChannel(value uint8, factor float64) uint8 {
	centered := float64(int(value) - 128)
	return clamp8(int(centered*factor + 128))
}

func clamp8(value int) uint8 {
	switch {
	case value < 0:
		return 0
	case value > 255:
		return 255
	default:
		return uint8(value)
	}
}
