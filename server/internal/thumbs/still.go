package thumbs

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"kallisti.chat/server/internal/imagesize"
)

// maxStill: a still's bytes at most; its data URL stays under the stream
// hub's 64 KiB.
const maxStill = 46 << 10

// StillArgs are ffmpeg's arguments for a JPEG of an H.264 keyframe from
// stdin (Annex B) to stdout, at most width by height.
func StillArgs(width, height int) []string {
	w, h := strconv.Itoa(width), strconv.Itoa(height)
	return []string{"-hide_banner", "-loglevel", "error", "-protocol_whitelist", "pipe", "-threads", "1",
		"-max_pixels", strconv.Itoa(imagesize.MaxPixels),
		"-f", "h264", "-i", "pipe:0", "-frames:v", "1", "-an", "-sn", "-dn",
		"-vf", "scale=w='min(" + w + ",iw)':h='min(" + h + ",ih)':force_original_aspect_ratio=decrease",
		"-c:v", "mjpeg", "-threads", "1", "-pix_fmt", "yuvj420p", "-q:v", "6", "-f", "mjpeg", "pipe:1"}
}

// Still makes a JPEG of an H.264 keyframe (Annex B, with its SPS and PPS)
// now, waiting for a free slot as Make does: stream previews (see whip).
func (m *Maker) Still(ctx context.Context, keyframe []byte, width, height int) ([]byte, error) {
	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	limit, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	out := &capped{max: maxStill}
	err := m.execIO(limit, StillArgs(width, height), keyframe, out)
	switch {
	case ctx.Err() != nil:
		return nil, ctx.Err()
	case limit.Err() != nil:
		return nil, fmt.Errorf("%w: ffmpeg: %w", ErrFailed, limit.Err())
	case errors.Is(err, errPassing):
		return nil, err
	case err != nil:
		return nil, fmt.Errorf("%w: %w", ErrFailed, err)
	case len(out.b) == 0 || out.over:
		return nil, fmt.Errorf("%w: ffmpeg wrote no picture, or a large one", ErrFailed)
	}
	return out.b, nil
}

// capped keeps up to max bytes and notes more.
type capped struct {
	b    []byte
	max  int
	over bool
}

func (c *capped) Write(p []byte) (int, error) {
	n := min(len(p), c.max-len(c.b))
	c.b = append(c.b, p[:n]...)
	c.over = c.over || n < len(p)
	return len(p), nil
}
