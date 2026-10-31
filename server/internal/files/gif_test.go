package files

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func TestGIFShowsTransparency(t *testing.T) {
	opaque := color.Palette{color.Black, color.White}
	clear := color.Palette{color.Black, color.Transparent}
	frame := func(p color.Palette, r image.Rectangle) *image.Paletted {
		img := image.NewPaletted(r, p)
		img.SetColorIndex(r.Min.X, r.Min.Y, 1)
		return img
	}
	full := image.Rect(0, 0, 10, 10)
	for _, c := range []struct {
		name string
		g    gif.GIF
		want bool
	}{
		{"opaque", gif.GIF{Image: []*image.Paletted{frame(opaque, full), frame(opaque, full)}, Delay: []int{5, 5}}, false},
		{"first frame transparent", gif.GIF{Image: []*image.Paletted{frame(clear, full), frame(opaque, full)}, Delay: []int{5, 5}}, true},
		// Later frames keep pixels of the one before: compression, not seen.
		{"later frame keeps the one before", gif.GIF{Image: []*image.Paletted{frame(opaque, full), frame(clear, full)}, Delay: []int{5, 5}, Disposal: []byte{gif.DisposalNone, gif.DisposalNone}}, false},
		{"later frame cleared to the background", gif.GIF{Image: []*image.Paletted{frame(opaque, full), frame(clear, full)}, Delay: []int{5, 5}, Disposal: []byte{gif.DisposalNone, gif.DisposalBackground}}, true},
		{"first frame smaller than the picture", gif.GIF{Image: []*image.Paletted{frame(opaque, image.Rect(2, 2, 10, 10))}, Delay: []int{5}, Config: image.Config{ColorModel: opaque, Width: 10, Height: 10}}, true},
	} {
		var b bytes.Buffer
		if err := gif.EncodeAll(&b, &c.g); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := gifShowsTransparency(bytes.NewReader(b.Bytes())); got != c.want {
			t.Errorf("%s: %v", c.name, got)
		}
	}
	// Not a GIF, or cut short: treated as transparent (no video).
	if !gifShowsTransparency(bytes.NewReader([]byte("GIF89a\x0a"))) {
		t.Error("cut short: not transparent")
	}
}
