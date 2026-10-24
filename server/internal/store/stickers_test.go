package store

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestStickers(t *testing.T) {
	s := open(t)
	t0 := time.UnixMilli(1_000_000).UTC()
	png := []byte("\x89PNG\r\n\x1a\nsticker")
	a, err := s.AddSticker(Sticker{Name: "wave", Description: "Hello!", Emoji: "👋", Type: "image/png", Data: png, Author: "alice"}, t0)
	if err != nil || a.ID == 0 || !a.CreatedAt.Equal(t0) {
		t.Fatalf("add: %+v %v", a, err)
	}
	if _, err := s.AddSticker(Sticker{Name: "wave", Type: "image/png", Data: png, Author: "bob"}, t0); !errors.Is(err, ErrNameTaken) {
		t.Fatalf("same name: %v", err)
	}
	s.AddSticker(Sticker{Name: "cat", Type: "image/gif", Data: []byte("GIF89a"), Author: "bob"}, t0)
	list, err := s.Stickers()
	if err != nil || len(list) != 2 || list[0].Name != "cat" || list[1].Name != "wave" || list[1].Data != nil || list[1].Emoji != "👋" || list[1].Description != "Hello!" {
		t.Fatalf("list: %+v %v", list, err)
	}
	got, err := s.StickerByID(a.ID)
	if err != nil || !bytes.Equal(got.Data, png) || got.Deleted {
		t.Fatalf("by id: %+v %v", got, err)
	}
	// Rename: name, description and emoji; a taken name is refused.
	if err := s.UpdateSticker(Sticker{ID: a.ID, Name: "hi", Description: "Hi!", Emoji: "🙂", Pack: "greetings"}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateSticker(Sticker{ID: a.ID, Name: "cat"}); !errors.Is(err, ErrNameTaken) {
		t.Fatalf("rename to a taken name: %v", err)
	}
	// Delete keeps the row and its name (old messages show it), not the
	// image; the name is free again, and the id is not used again.
	if err := s.DeleteSticker(a.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteSticker(a.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted twice: %v", err)
	}
	got, err = s.StickerByID(a.ID)
	if err != nil || !got.Deleted || got.Data != nil || got.Name != "hi" || got.Pack != "greetings" {
		t.Fatalf("deleted: %+v %v", got, err)
	}
	if err := s.UpdateSticker(Sticker{ID: a.ID, Name: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("rename a deleted one: %v", err)
	}
	if list, _ := s.Stickers(); len(list) != 1 {
		t.Fatalf("list after delete: %+v", list)
	}
	b, err := s.AddSticker(Sticker{Name: "hi", Type: "image/png", Data: png, Author: "bob"}, t0)
	if err != nil || b.ID <= a.ID {
		t.Fatalf("the name again: %+v %v", b, err)
	}
	if _, err := s.StickerByID(9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown: %v", err)
	}
}

func TestStickerLimit(t *testing.T) {
	s := open(t)
	var last Sticker
	for i := range MaxStickers {
		var err error
		if last, err = s.AddSticker(Sticker{Name: fmt.Sprintf("s%d", i), Type: "image/png", Data: []byte("x"), Author: "a"}, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.AddSticker(Sticker{Name: "more", Type: "image/png", Data: []byte("x"), Author: "a"}, time.Now()); !errors.Is(err, ErrTooManyStickers) {
		t.Fatalf("over the limit: %v", err)
	}
	s.DeleteSticker(last.ID) // deleted ones do not count
	if _, err := s.AddSticker(Sticker{Name: "more", Type: "image/png", Data: []byte("x"), Author: "a"}, time.Now()); err != nil {
		t.Fatal(err)
	}
}
