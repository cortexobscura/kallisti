package whip

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"kallisti.chat/server/internal/stream"
)

func TestPreviewDue(t *testing.T) {
	now := time.Now()
	for _, c := range []struct {
		last time.Time
		busy bool
		want bool
	}{
		{time.Time{}, false, true}, // the first keyframe
		{now.Add(-previewEvery), false, true},
		{now.Add(-previewEvery + time.Second), false, false},
		{time.Time{}, true, false}, // one at a time
	} {
		if got := previewDue(c.last, now, c.busy); got != c.want {
			t.Errorf("last %v busy %v: %v", c.last, c.busy, got)
		}
	}
}

// A keyframe becomes the stream's preview (as a JPEG data URL), once per
// previewEvery.
func TestKeyframesBecomePreviews(t *testing.T) {
	h, _, _ := hubWithViewer()
	var b bytes.Buffer
	jpeg.Encode(&b, image.NewGray(image.Rect(0, 0, 32, 18)), nil)
	var calls atomic.Int32
	got := make(chan string, 4)
	h.OnPreview = func(i stream.Info) { got <- i.Preview }
	f := newFeed(h, 1, false, 1<<20, stream.IngestHooks{Keyframe: func() {}, Ended: func() {}})
	f.still = func(_ context.Context, key []byte) ([]byte, error) {
		calls.Add(1)
		if !bytes.Contains(key, []byte{0, 0, 0, 1}) {
			t.Error("not the keyframe's Annex B")
		}
		return b.Bytes(), nil
	}
	v := newVideoIn()
	frames := clipFrames(t)
	for i, p := range packets(append(frames, frames...), 0, 0) {
		for _, pic := range v.push(p, int64(i)*1000) {
			f.picture(pic)
		}
	}
	select {
	case p := <-got:
		if !strings.HasPrefix(p, "data:image/jpeg;base64,") {
			t.Fatalf("preview %.40q", p)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no preview")
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("%d stills for keyframes within %v", n, previewEvery)
	}
}

// A stop ends a preview in the making, and it never lands.
func TestStopEndsAPreview(t *testing.T) {
	h, _, _ := hubWithViewer()
	landed := make(chan struct{}, 1)
	h.OnPreview = func(stream.Info) { landed <- struct{}{} }
	f := newFeed(h, 1, false, 1<<20, stream.IngestHooks{Keyframe: func() {}, Ended: func() {}})
	started, ended := make(chan struct{}), make(chan error, 1)
	f.still = func(ctx context.Context, _ []byte) ([]byte, error) {
		close(started)
		<-ctx.Done()
		ended <- ctx.Err()
		return []byte{0xff, 0xd8}, nil
	}
	v := newVideoIn()
	for i, p := range packets(clipFrames(t), 0, 0) {
		for _, pic := range v.push(p, int64(i)*1000) {
			f.picture(pic)
		}
	}
	<-started
	f.stop()
	if err := <-ended; err == nil {
		t.Fatal("not cancelled")
	}
	select {
	case <-landed:
		t.Fatal("a preview after the stop")
	case <-time.After(200 * time.Millisecond):
	}
}
