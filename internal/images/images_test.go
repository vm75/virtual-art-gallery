package images

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
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

func TestValidateConfigRejectsExcessiveDecodedPixels(t *testing.T) {
	if err := validateConfig(image.Config{Width: 8000, Height: 6000}, "png"); err == nil {
		t.Fatal("excessive decoded pixel count accepted")
	}
	if err := validateConfig(image.Config{Width: 8000, Height: 5000}, "png"); err != nil {
		t.Fatalf("pixel boundary rejected: %v", err)
	}
}

func TestProcessBoundsPortraitLandscapeAndSmallImages(t *testing.T) {
	for _, test := range []struct {
		name          string
		width, height int
	}{
		{name: "portrait", width: 100, height: 10000},
		{name: "landscape", width: 10000, height: 100},
		{name: "small", width: 100, height: 50},
	} {
		t.Run(test.name, func(t *testing.T) {
			src := image.NewRGBA(image.Rect(0, 0, test.width, test.height))
			var input bytes.Buffer
			if err := png.Encode(&input, src); err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			result, err := (Pipeline{Root: root}).Process(&input, 1)
			if err != nil {
				t.Fatal(err)
			}
			for label, relative := range map[string]string{
				"thumbnail": result.Thumbnail,
				"medium":    result.Medium,
				"museum":    result.Museum,
				"large":     result.Large,
			} {
				file, err := os.Open(filepath.Join(root, relative))
				if err != nil {
					t.Fatal(err)
				}
				config, _, err := image.DecodeConfig(file)
				file.Close()
				if err != nil {
					t.Fatal(err)
				}
				limits := derivativeSizes[label]
				if config.Width > limits.Width || config.Height > limits.Height {
					t.Fatalf("%s derivative = %dx%d, limits %dx%d", label, config.Width, config.Height, limits.Width, limits.Height)
				}
				if test.width <= limits.Width && test.height <= limits.Height {
					if config.Width != test.width || config.Height != test.height {
						t.Fatalf("%s unexpectedly upscaled small image to %dx%d", label, config.Width, config.Height)
					}
					continue
				}
				outputAspect := float64(config.Width) / float64(config.Height)
				sourceAspect := float64(test.width) / float64(test.height)
				if math.Abs(outputAspect/sourceAspect-1) > 0.1 {
					t.Fatalf("%s aspect ratio changed: %dx%d from %dx%d", label, config.Width, config.Height, test.width, test.height)
				}
			}
		})
	}
}

func TestFitUsesBilinearSampling(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3, 1))
	src.SetRGBA(0, 0, color.RGBA{A: 255})
	src.SetRGBA(1, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(2, 0, color.RGBA{A: 255})
	resized := fit(src, derivativeSize{Width: 2, Height: 1})
	for x := 0; x < 2; x++ {
		red, _, _, _ := resized.At(x, 0).RGBA()
		if red == 0 || red == 0xffff {
			t.Fatalf("pixel %d was not interpolated: %d", x, red)
		}
	}
}
