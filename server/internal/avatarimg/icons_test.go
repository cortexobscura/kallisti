package avatarimg

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestIcons(t *testing.T) {
	// A wide picture, its top quarter transparent.
	src := halves(600, 300)
	for i := 3; i < 600*75*4; i += 4 {
		src.Pix[i] = 0
	}
	icons, err := Icons(pngOf(t, src), []int{32, 512})
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{32, 512} {
		img, err := png.Decode(bytes.NewReader(icons[size]))
		if err != nil {
			t.Fatalf("%d: not a PNG: %v", size, err)
		}
		if b := img.Bounds(); b.Dx() != size || b.Dy() != size {
			t.Fatalf("%d: %v", size, b)
		}
		// The middle square: red left, blue right; transparency kept.
		wantColor(t, "left", img, size/4, size*3/4, red)
		wantColor(t, "right", img, size*3/4, size*3/4, blue)
		if _, _, _, a := img.At(size/2, size/16).RGBA(); a > 0x4000 {
			t.Errorf("%d: the transparent top is %x", size, a)
		}
	}
	if _, err := Icons([]byte("not a picture"), []int{32}); err == nil {
		t.Fatal("no error for a broken picture")
	}
	huge := image.NewNRGBA(image.Rect(0, 0, MaxSide+1, 1))
	if _, err := Icons(pngOf(t, huge), []int{32}); err == nil {
		t.Fatal("no error for a huge picture")
	}
}
