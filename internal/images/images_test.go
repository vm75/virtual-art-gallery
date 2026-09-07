package images

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessCreatesBoundedDerivativesAndOriginal(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1000, 500))
	for y := 0; y < 500; y++ {
		for x := 0; x < 1000; x++ {
			src.Set(x, y, color.RGBA{R: 220, A: 255})
		}
	}
	var input bytes.Buffer
	if err := png.Encode(&input, src); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	result, err := (Pipeline{Root: root, MaxSize: 2 << 20}).Process(&input, 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.Width != 1000 || result.Height != 500 {
		t.Fatalf("dimensions = %dx%d", result.Width, result.Height)
	}
	for _, relative := range []string{result.Original, result.Thumbnail, result.Medium, result.Museum, result.Large} {
		if filepath.IsAbs(relative) || filepath.Base(relative) == "" {
			t.Fatalf("unsafe path %q", relative)
		}
		if _, err := os.Stat(filepath.Join(root, relative)); err != nil {
			t.Fatalf("missing %s: %v", relative, err)
		}
	}
}

func TestProcessRejectsInvalidAndOversize(t *testing.T) {
	if _, err := (Pipeline{Root: t.TempDir(), MaxSize: 4}).Process(bytes.NewReader([]byte("not image")), 1); err == nil {
		t.Fatal("expected invalid image error")
	}
	if _, err := (Pipeline{Root: t.TempDir(), MaxSize: 4}).Process(bytes.NewReader(bytes.Repeat([]byte("x"), 5)), 1); err == nil {
		t.Fatal("expected size error")
	}
}
