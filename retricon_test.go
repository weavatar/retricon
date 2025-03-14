package retricon

import (
	"image"
	"image/color"
	"testing"
)

func TestDefaultStyleRetriconGeneration(t *testing.T) {
	img, err := New("test")

	if err != nil {
		t.Fatalf("Failed to generate default retricon: %v", err)
	}

	if img == nil {
		t.Fatal("Generated image is nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Error("Generated image has invalid dimensions")
	}
}

func TestAllPredefinedStyles(t *testing.T) {
	styles := []Style{
		Default,
		Github,
		Gravatar,
		Mono,
		Mosaic,
		Mini,
		Window,
	}

	for _, style := range styles {
		img, err := New("test", style)

		if err != nil {
			t.Errorf("Failed to generate retricon with style %s: %v", style, err)
			continue
		}

		if img == nil {
			t.Errorf("Generated image for style %s is nil", style)
		}
	}
}

func TestInvalidStyle(t *testing.T) {
	_, err := New("test", "invalid-style")

	if err == nil {
		t.Error("Expected error for invalid style, but got nil")
	}
}

func TestCustomOptions(t *testing.T) {
	opts := Options{
		Tiles:         8,
		TileSize:      10,
		TileColor:     "FF0000", // Red
		BgColor:       "000000", // Black
		TilePadding:   2,
		ImagePadding:  5,
		Size:          500,
		VerticalSym:   true,
		HorizontalSym: false,
	}

	img, err := NewWithOptions("test", opts)

	if err != nil {
		t.Fatalf("Failed to generate retricon with custom options: %v", err)
	}

	expectedSize := 499
	if img.Bounds().Dx() != expectedSize || img.Bounds().Dy() != expectedSize {
		t.Errorf("Expected image size to be %d, got %dx%d",
			expectedSize, img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestSymmetryOptions(t *testing.T) {
	testCases := []struct {
		name          string
		verticalSym   bool
		horizontalSym bool
	}{
		{"no-symmetry", false, false},
		{"vertical-symmetry", true, false},
		{"horizontal-symmetry", false, true},
		{"both-symmetries", true, true},
	}

	for _, tc := range testCases {
		opts := Options{
			Tiles:         6,
			TileSize:      5,
			VerticalSym:   tc.verticalSym,
			HorizontalSym: tc.horizontalSym,
		}

		img, err := NewWithOptions(tc.name, opts)

		if err != nil {
			t.Errorf("Failed to generate retricon with %s: %v", tc.name, err)
			continue
		}

		if img == nil {
			t.Errorf("Generated image for %s is nil", tc.name)
		}
	}
}

func TestColorFormats(t *testing.T) {
	testCases := []struct {
		name      string
		tileColor interface{}
		bgColor   interface{}
	}{
		{"hex-colors", "FF0000", "00FF00"},
		{"rgba-colors", color.RGBA{R: 255, G: 0, B: 0, A: 255}, color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"color-indices", 0, 1},
		{"uint8-slice", []uint8{255, 0, 0, 255}, []uint8{0, 255, 0, 255}},
	}

	for _, tc := range testCases {
		opts := Options{
			Tiles:     4,
			TileSize:  5,
			TileColor: tc.tileColor,
			BgColor:   tc.bgColor,
		}

		img, err := NewWithOptions(tc.name, opts)

		if err != nil {
			t.Errorf("Failed with color format %s: %v", tc.name, err)
			continue
		}

		if img == nil {
			t.Errorf("Generated image for %s is nil", tc.name)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	// Empty name
	img1, err1 := New("")
	if err1 != nil {
		t.Errorf("Failed to generate retricon with empty name: %v", err1)
	}
	if img1 == nil {
		t.Error("Generated image for empty name is nil")
	}

	// Very long name
	longName := string(make([]byte, 10000))
	img2, err2 := New(longName)
	if err2 != nil {
		t.Errorf("Failed to generate retricon with very long name: %v", err2)
	}
	if img2 == nil {
		t.Error("Generated image for long name is nil")
	}
}

func TestMustFunctions(t *testing.T) {
	// Test MustNew
	img1 := MustNew("test")
	if img1 == nil {
		t.Error("MustNew returned nil image")
	}

	// Test MustNewWithOptions
	opts := Options{
		Tiles:    3,
		TileSize: 10,
	}
	img2 := MustNewWithOptions("test", opts)
	if img2 == nil {
		t.Error("MustNewWithOptions returned nil image")
	}

	// Test panics with defer/recover
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew with invalid style did not panic")
		}
	}()
	_ = MustNew("test", "invalid-style")
}

func TestSameNameSameImage(t *testing.T) {
	img1, _ := New("test")
	img2, _ := New("test")

	equal := compareImages(img1, img2)
	if !equal {
		t.Error("Images generated with the same name are different")
	}
}

func TestDifferentNamesDifferentImages(t *testing.T) {
	img1, _ := New("test1")
	img2, _ := New("test2")

	equal := compareImages(img1, img2)
	if equal {
		t.Error("Images generated with different names are identical")
	}
}

// Helper function to compare images
func compareImages(img1, img2 image.Image) bool {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		return false
	}

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				return false
			}
		}
	}
	return true
}
