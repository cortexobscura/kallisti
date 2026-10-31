package sidecar

import (
	"encoding/json"
	"log/slog"

	"kallisti.chat/server/internal/mumble"
)

// channelList is a registered user's own channel list: hidden channels
// and an own order, of text channels (their ids) and voice channels (their
// Mumble ids). The client applies it; the server keeps it for the user's
// devices and checks only its size.
type channelList struct {
	Hidden listIDs `json:"hidden"`
	Order  listIDs `json:"order"`
}

type listIDs struct {
	Text  []int64  `json:"text"`
	Voice []uint32 `json:"voice"`
}

// maxListIDs: per list, more than any server has channels.
const maxListIDs = 1000

func (p channelList) valid() bool {
	for _, n := range []int{len(p.Hidden.Text), len(p.Hidden.Voice), len(p.Order.Text), len(p.Order.Voice)} {
		if n > maxListIDs {
			return false
		}
	}
	return true
}

// lists: empty lists, never null, on the wire.
func (p channelList) lists() channelList {
	for _, l := range []*[]int64{&p.Hidden.Text, &p.Order.Text} {
		if *l == nil {
			*l = []int64{}
		}
	}
	for _, l := range []*[]uint32{&p.Hidden.Voice, &p.Order.Voice} {
		if *l == nil {
			*l = []uint32{}
		}
	}
	return p
}

// channelListOf: a registered user's kept list (empty if none).
func (h *Hub) channelListOf(user uint32) channelList {
	var p channelList
	raw, err := h.store.ChannelList(user)
	if err != nil {
		slog.Error("channel list", "err", err)
	} else if raw != "" {
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			slog.Error("channel list", "err", err)
		}
	}
	return p.lists()
}

// setChannelList keeps a user's list and sends it to all its connections
// (its other devices).
func (c *Conn) setChannelList(id mumble.Identity, p *channelList) {
	switch {
	case !id.Registered:
		c.sendError("join with an account (or a guest account) to keep your channel list")
		return
	case p == nil || !p.valid():
		c.sendError("not a channel list")
		return
	}
	list := p.lists()
	raw, _ := json.Marshal(list)
	if err := c.hub.store.SetChannelList(id.UserID, string(raw)); err != nil {
		slog.Error("keep channel list", "err", err)
		c.sendError("cannot keep your channel list")
		return
	}
	c.hub.eachReady(func(other *Conn, oid mumble.Identity) {
		if oid.Registered && oid.UserID == id.UserID {
			other.send(map[string]any{"type": "channel-list", "list": list})
		}
	})
}
