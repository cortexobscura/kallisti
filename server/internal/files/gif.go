package files

import (
	"bufio"
	"encoding/binary"
	"io"
)

// gifShowsTransparency: a GIF shows see-through pixels (an MP4 has none: a
// box), read from its block headers only. The first frame with
// transparency or not covering the picture shows it; so does a later
// frame with transparency cleared to the background (disposal 2). Later
// frames that keep the one before use transparency to compress, unseen.
// A file that is not a whole GIF counts as transparent (no video).
func gifShowsTransparency(r io.Reader) bool {
	br := bufio.NewReader(r)
	var head [13]byte // header, logical screen descriptor
	if _, err := io.ReadFull(br, head[:]); err != nil || string(head[:3]) != "GIF" {
		return true
	}
	width, height := binary.LittleEndian.Uint16(head[6:]), binary.LittleEndian.Uint16(head[8:])
	if head[10]&0x80 != 0 && skip(br, 3<<(head[10]&7+1)) != nil {
		return true
	}
	gce, first := byte(0), true
	for {
		b, err := br.ReadByte()
		if err != nil {
			return true
		}
		switch b {
		case 0x21: // extension
			label, err := br.ReadByte()
			if err != nil {
				return true
			}
			if label == 0xf9 { // graphic control: the next frame's
				var ext [6]byte // size, packed, delay, index, terminator
				if _, err := io.ReadFull(br, ext[:]); err != nil {
					return true
				}
				gce = ext[1]
				continue
			}
			if skipBlocks(br) != nil {
				return true
			}
		case 0x2c: // a frame
			var desc [9]byte
			if _, err := io.ReadFull(br, desc[:]); err != nil {
				return true
			}
			transparent, disposal := gce&1 != 0, gce>>2&7
			left, top := binary.LittleEndian.Uint16(desc[0:]), binary.LittleEndian.Uint16(desc[2:])
			w, h := binary.LittleEndian.Uint16(desc[4:]), binary.LittleEndian.Uint16(desc[6:])
			covers := left == 0 && top == 0 && w >= width && h >= height
			if first && (transparent || !covers) || transparent && disposal == 2 {
				return true
			}
			first, gce = false, 0
			if desc[8]&0x80 != 0 && skip(br, 3<<(desc[8]&7+1)) != nil {
				return true
			}
			if _, err := br.ReadByte(); err != nil { // LZW code size
				return true
			}
			if skipBlocks(br) != nil {
				return true
			}
		case 0x3b: // the end
			return first
		default:
			return true
		}
	}
}

func skip(br *bufio.Reader, n int) error {
	_, err := br.Discard(n)
	return err
}

// skipBlocks skips data sub-blocks up to the empty one.
func skipBlocks(br *bufio.Reader) error {
	for {
		n, err := br.ReadByte()
		if err != nil || n == 0 {
			return err
		}
		if err := skip(br, int(n)); err != nil {
			return err
		}
	}
}
