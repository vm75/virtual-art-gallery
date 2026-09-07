package images

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	DefaultMaxUpload = 20 << 20
	MaxDimension     = 16000
)

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
	if config.Width < 1 || config.Height < 1 || config.Width > MaxDimension || config.Height > MaxDimension {
		return Result{}, fmt.Errorf("image dimensions are outside supported limits")
	}
	if format != "jpeg" && format != "png" && format != "gif" {
		return Result{}, fmt.Errorf("unsupported image format %q", format)
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
	for label, width := range map[string]int{"thumbnail": 480, "medium": 1200, "museum": 2048, "large": 2400} {
		path := filepath.Join(temp, label+".jpg")
		if err := writeJPEG(path, fit(decoded, width)); err != nil {
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

func decode(data []byte) (image.Image, error) {
	decoded, _, err := image.Decode(bytes.NewReader(data))
	return decoded, err
}

func fit(src image.Image, maxWidth int) image.Image {
	b := src.Bounds()
	if b.Dx() <= maxWidth {
		return src
	}
	h := b.Dy() * maxWidth / b.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, maxWidth, h))
	for y := 0; y < h; y++ {
		for x := 0; x < maxWidth; x++ {
			dst.Set(x, y, src.At(b.Min.X+x*b.Dx()/maxWidth, b.Min.Y+y*b.Dy()/h))
		}
	}
	return dst
}

func writeJPEG(path string, img image.Image) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 88})
}
