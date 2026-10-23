package thumbs

import (
	"bytes"
	"context"
	"errors"
	"image/jpeg"
	"os/exec"
	"slices"
	"testing"
	"time"
)

func TestStillArgs(t *testing.T) {
	a := StillArgs(320, 1024)
	for _, want := range [][]string{
		{"-protocol_whitelist", "pipe"},
		{"-f", "h264", "-i", "pipe:0"},
		{"-frames:v", "1"},
		{"-c:v", "mjpeg"},
		{"-f", "mjpeg", "pipe:1"},
	} {
		if !containsRun(a, want) {
			t.Errorf("args %v lack %v", a, want)
		}
	}
	if slices.Contains(a, "-nostdin") {
		t.Error("-nostdin: the keyframe comes on stdin")
	}
}

func containsRun(a, run []string) bool {
	for i := range a {
		if len(a)-i >= len(run) && slices.Equal(a[i:i+len(run)], run) {
			return true
		}
	}
	return false
}

func TestStillPipes(t *testing.T) {
	// The fake ffmpeg echoes its stdin: the keyframe goes in, the picture comes out.
	m := New(t.Context(), fakeFFmpeg(t, `cat`), 1, 1)
	got, err := m.Still(t.Context(), []byte("keyframe"), 320, 1024)
	if err != nil || string(got) != "keyframe" {
		t.Fatalf("still %q %v", got, err)
	}
	for name, body := range map[string]string{"fails": `exit 1`, "nothing": `:`, "huge": `head -c 70000 /dev/zero`} {
		m := New(t.Context(), fakeFFmpeg(t, body), 1, 1)
		if _, err := m.Still(t.Context(), []byte("k"), 320, 1024); !errors.Is(err, ErrFailed) {
			t.Errorf("%s: %v", name, err)
		}
	}
	slow := New(t.Context(), fakeFFmpeg(t, `exec sleep 30`), 1, 1)
	slow.timeout = 200 * time.Millisecond
	if _, err := slow.Still(t.Context(), []byte("k"), 320, 1024); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("slow: %v", err)
	}
}

// With a real ffmpeg (skipped without one): a keyframe from x264 becomes a
// JPEG of the asked width.
func TestRealFFmpegStill(t *testing.T) {
	tool, err := Find("")
	if err != nil {
		t.Skip("no ffmpeg:", err)
	}
	h264, err := exec.Command(tool.Command[len(tool.Command)-1], "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=1280x720:rate=30",
		"-frames:v", "1", "-c:v", "libx264", "-profile:v", "baseline", "-pix_fmt", "yuv420p", "-f", "h264", "pipe:1").Output()
	if err != nil {
		t.Skip("no libx264:", err)
	}
	m := New(t.Context(), tool, 1, 1)
	pic, err := m.Still(t.Context(), h264, 320, 1024)
	if err != nil {
		t.Fatal(err)
	}
	c, err := jpeg.DecodeConfig(bytes.NewReader(pic))
	if err != nil || c.Width != 320 || c.Height != 180 {
		t.Fatalf("picture %+v %v", c, err)
	}
}
