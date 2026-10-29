package sidecar

import (
	"regexp"
	"slices"
	"strings"

	"kallisti.chat/server/internal/mumble"
)

// Text channel categories (built-in server): one level, like Discord's.
// Each has a Mumble channel under the mirrors' parent that holds its
// channels' mirrors, so a category's moderators (its "channelmod" group)
// moderate all its channels and its time-outs hold there. Staff manage
// categories and move channels between them; a category's moderators
// create, rename, order and delete the channels in it.

type wireCategory struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Mumble uint32 `json:"mumble,omitempty"` // holds its channels' mirrors
}

// Category names: words of letters, digits and the other characters of
// Mumble channel names but "#" (mirrors start with it).
var categoryNameChars = regexp.MustCompile(`^[\p{L}\p{N}\p{M} _\-=\[\]{}()@|]{1,32}$`)

func categoryName(raw string) (string, bool) {
	name := strings.Join(strings.Fields(raw), " ")
	return name, categoryNameChars.MatchString(name)
}

// mayManage: staff change all text channels and categories. A channel's
// moderators set its topic and slow mode; a category's moderators also
// create (not private ones), rename, order and delete its channels (those
// they see).
func (h *Hub) mayManage(id mumble.Identity, req request) bool {
	if h.voice == nil {
		return false
	}
	if h.voice.MayManage(id.Session) {
		return true
	}
	own := h.mirrorOf(req.Channel) != 0 && h.mayModerate(id, chat{id: req.Channel})
	switch req.Type {
	case "set-topic", "set-slow-mode":
		return own
	case "rename-channel", "delete-channel":
		return own && h.managesCategory(id, h.categoryOf(req.Channel))
	case "create-channel":
		return !req.Private && h.managesCategory(id, req.Category)
	case "order-channels":
		if len(req.Channels) == 0 {
			return false
		}
		cat := h.categoryOf(req.Channels[0])
		return !slices.ContainsFunc(req.Channels, func(ch int64) bool { return h.categoryOf(ch) != cat }) && h.managesCategory(id, cat)
	}
	return false
}

// managesCategory: the user moderates the category (0: none).
func (h *Hub) managesCategory(id mumble.Identity, category int64) bool {
	m := h.categoryMirror(category)
	return m != 0 && h.voice.ModeratesIn(id.Session, m)
}

// categoryOf is a text channel's category (0: none, or no such channel).
func (h *Hub) categoryOf(channel int64) int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i := slices.IndexFunc(h.channels, func(c wireChannel) bool { return c.ID == channel }); i >= 0 {
		return h.channels[i].Category
	}
	return 0
}

// categoryMirror is a category's Mumble channel (0: none).
func (h *Hub) categoryMirror(category int64) uint32 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i := slices.IndexFunc(h.categories, func(c wireCategory) bool { return c.ID == category }); category != 0 && i >= 0 {
		return h.categories[i].Mumble
	}
	return 0
}

// categoriesFor: the categories with a channel the user reads (of
// channels), and those it moderates (staff all); a list, never null in
// JSON.
func (h *Hub) categoriesFor(id mumble.Identity, channels []wireChannel) []wireCategory {
	h.mu.Lock()
	all := append([]wireCategory{}, h.categories...)
	h.mu.Unlock()
	return slices.DeleteFunc(all, func(c wireCategory) bool {
		reads := slices.ContainsFunc(channels, func(ch wireChannel) bool { return ch.Category == c.ID })
		return !reads && (h.voice == nil || c.Mumble == 0 || !h.voice.ModeratesIn(id.Session, c.Mumble))
	})
}

// channelsEvent: the text channels the user reads, and their categories.
func (h *Hub) channelsEvent(id mumble.Identity) map[string]any {
	channels := h.channelsFor(id)
	return map[string]any{"type": "channels", "channels": channels, "categories": h.categoriesFor(id, channels), "threads": h.threadsFor(id)}
}

// exposesChannel: moving the text channel to the category (0: none) would
// show it to everyone: a public channel out of a private category to a
// public place (its history would be everyone's).
func (h *Hub) exposesChannel(channel, to int64) bool {
	if h.voice == nil {
		return false
	}
	h.mu.Lock()
	i := slices.IndexFunc(h.channels, func(c wireChannel) bool { return c.ID == channel })
	var ch wireChannel
	if i >= 0 {
		ch = h.channels[i]
	}
	from, dest := h.categoryMumble(ch.Category), h.categoryMumble(to)
	h.mu.Unlock()
	return i >= 0 && from != 0 && h.isPrivate(from) && !(ch.Mumble != 0 && h.isPrivate(ch.Mumble)) && !(dest != 0 && h.isPrivate(dest))
}

// exposes: deleting the category would move such a channel out of it.
func (h *Hub) exposes(category, to int64) bool {
	h.mu.Lock()
	var ids []int64
	for _, c := range h.channels {
		if c.Category == category {
			ids = append(ids, c.ID)
		}
	}
	h.mu.Unlock()
	return slices.ContainsFunc(ids, func(id int64) bool { return h.exposesChannel(id, to) })
}

// categoryMumble: a category's Mumble channel (0: none). Under mu.
func (h *Hub) categoryMumble(id int64) uint32 {
	if i := slices.IndexFunc(h.categories, func(c wireCategory) bool { return c.ID == id }); id != 0 && i >= 0 {
		return h.categories[i].Mumble
	}
	return 0
}
