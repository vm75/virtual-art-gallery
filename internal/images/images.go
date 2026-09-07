package images

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"
	"os"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	DefaultMaxUpload = 20 << 20
	MaxDimension     = 16000
	MaxDecodedPixels = 40_000_000
)

type derivativeSize struct{ Width, Height int }

var derivativeSizes = map[string]derivativeSize{
	"thumbnail": {Width: 480, Height: 480},
	"medium":    {Width: 1200, Height: 1200},
	"museum":    {Width: 2048, Height: 2048},
	"large":     {Width: 2400, Height: 2400},
}

type Result struct {
	Original, Thumbnail, Medium, Museum, Large string
	Width, Height                              int
}

type Pipeline struct {
	Root    string
	MaxSize int64
}

func (p Pipeline) Process(input io.Reader, artworkID int64) (Result, error) {
	if artworkID < 1 || p.Root == "" {
		return Result{}, fmt.Errorf("invalid image storage configuration")
	}
	maxSize := p.MaxSize
	if maxSize <= 0 {
		maxSize = DefaultMaxUpload
	}
	limited := io.LimitReader(input, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return Result{}, fmt.Errorf("read image: %w", err)
	}
	if int64(len(data)) > maxSize {
		return Result{}, fmt.Errorf("image exceeds %d byte limit", maxSize)
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Result{}, fmt.Errorf("decode image: %w", err)
	}
	if err := validateConfig(config, format); err != nil {
		return Result{}, err
	}
	decoded, err := decode(data)
	if err != nil {
		return Result{}, fmt.Errorf("decode image: %w", err)
	}
	nameBytes := make([]byte, 12)
	if _, err := rand.Read(nameBytes); err != nil {
		return Result{}, fmt.Errorf("generate image name: %w", err)
	}
	base := hex.EncodeToString(nameBytes)
	dir := filepath.Join(p.Root, "images", fmt.Sprintf("%d-%s", artworkID, base))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return Result{}, fmt.Errorf("create image directory: %w", err)
	}
	temp := filepath.Join(dir, ".processing")
	if err := os.Mkdir(temp, 0o750); err != nil {
		return Result{}, fmt.Errorf("create temporary image directory: %w", err)
	}
	published := false
	defer func() {
		_ = os.RemoveAll(temp)
		if !published {
			_ = os.RemoveAll(dir)
		}
	}()
	paths := map[string]string{}
	for label, size := range derivativeSizes {
		path := filepath.Join(temp, label+".jpg")
		if err := writeJPEG(path, fit(decoded, size)); err != nil {
			return Result{}, fmt.Errorf("write %s derivative: %w", label, err)
		}
		paths[label] = filepath.Join("images", filepath.Base(dir), label+".jpg")
	}
	ext := filepath.Ext("." + format)
	originalPath := filepath.Join(temp, "original"+ext)
	if err := os.WriteFile(originalPath, data, 0o640); err != nil {
		return Result{}, fmt.Errorf("write original: %w", err)
	}
	for _, label := range []string{"thumbnail", "medium", "museum", "large"} {
		if err := os.Rename(filepath.Join(temp, label+".jpg"), filepath.Join(dir, label+".jpg")); err != nil {
			return Result{}, fmt.Errorf("publish %s derivative: %w", label, err)
		}
	}
	if err := os.Rename(originalPath, filepath.Join(dir, "original"+ext)); err != nil {
		return Result{}, fmt.Errorf("publish original: %w", err)
	}
	published = true
	return Result{Original: filepath.Join("images", filepath.Base(dir), "original"+ext), Thumbnail: paths["thumbnail"], Medium: paths["medium"], Museum: paths["museum"], Large: paths["large"], Width: config.Width, Height: config.Height}, nil
}

func validateConfig(config image.Config, format string) error {
	if config.Width < 1 || config.Height < 1 || config.Width > MaxDimension || config.Height > MaxDimension {
		return fmt.Errorf("image dimensions are outside supported limits")
	}
	if int64(config.Width)*int64(config.Height) > MaxDecodedPixels {
		return fmt.Errorf("image exceeds %d decoded pixel limit", MaxDecodedPixels)
	}
	if format != "jpeg" && format != "png" && format != "gif" {
		return fmt.Errorf("unsupported image format %q", format)
	}
	return nil
}

func decode(data []byte) (image.Image, error) {
	decoded, _, err := image.Decode(bytes.NewReader(data))
	return decoded, err
}

func fit(src image.Image, limits derivativeSize) image.Image {
	b := src.Bounds()
	width, height := b.Dx(), b.Dy()
	if width <= limits.Width && height <= limits.Height {
		return src
	}
	scale := math.Min(float64(limits.Width)/float64(width), float64(limits.Height)/float64(height))
	resizedWidth := max(1, int(math.Round(float64(width)*scale)))
	resizedHeight := max(1, int(math.Round(float64(height)*scale)))
	dst := image.NewRGBA(image.Rect(0, 0, resizedWidth, resizedHeight))
	for y := 0; y < resizedHeight; y++ {
		sy := (float64(y)+0.5)*float64(height)/float64(resizedHeight) - 0.5
		y0, y1, fy := interpolationCoordinates(sy, height)
		for x := 0; x < resizedWidth; x++ {
			sx := (float64(x)+0.5)*float64(width)/float64(resizedWidth) - 0.5
			x0, x1, fx := interpolationCoordinates(sx, width)
			c00 := color.RGBAModel.Convert(src.At(b.Min.X+x0, b.Min.Y+y0)).(color.RGBA)
			c10 := color.RGBAModel.Convert(src.At(b.Min.X+x1, b.Min.Y+y0)).(color.RGBA)
			c01 := color.RGBAModel.Convert(src.At(b.Min.X+x0, b.Min.Y+y1)).(color.RGBA)
			c11 := color.RGBAModel.Convert(src.At(b.Min.X+x1, b.Min.Y+y1)).(color.RGBA)
			dst.SetRGBA(x, y, color.RGBA{
				R: bilinear(c00.R, c10.R, c01.R, c11.R, fx, fy),
				G: bilinear(c00.G, c10.G, c01.G, c11.G, fx, fy),
				B: bilinear(c00.B, c10.B, c01.B, c11.B, fx, fy),
				A: bilinear(c00.A, c10.A, c01.A, c11.A, fx, fy),
			})
		}
	}
	return dst
}

func interpolationCoordinates(value float64, length int) (int, int, float64) {
	if value <= 0 {
		return 0, 0, 0
	}
	if value >= float64(length-1) {
		return length - 1, length - 1, 0
	}
	first := int(math.Floor(value))
	return first, first + 1, value - float64(first)
}

func bilinear(a, b, c, d uint8, x, y float64) uint8 {
	top := float64(a)*(1-x) + float64(b)*x
	bottom := float64(c)*(1-x) + float64(d)*x
	return uint8(math.Round(top*(1-y) + bottom*y))
}

func writeJPEG(path string, img image.Image) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 88})
}
