package main

import (
	"fmt"
	"strings"
	"time"
)

// packet is one voice packet as heard: its frame number (10 ms units) and
// terminator flag.
type packet struct {
	frame      uint64
	terminator bool
	at         time.Time
}

// spurt is one talk spurt and what is wrong with it.
type spurt struct {
	first, last     uint64
	packets         int
	terminators     int
	gaps, repeats   int  // frame number steps above 2, or not upward
	afterTerminator bool // began right after a terminator (an extra packet)
	open            bool // still running at the end (no terminator yet)
	// Pauses in it with the frame numbers going on (see maxStall), and the
	// longest: the sender held its packets back for a moment.
	stalls       int
	longestStall time.Duration
}

type report struct{ spurts []spurt }

const (
	frameStep  = 2 // 20 ms packets
	spurtPause = 300 * time.Millisecond
	extraGap   = 100 * time.Millisecond
	// A pause up to this long whose next packet goes on with the frame
	// numbers is a stall, not a spurt end: a slow sender (Safari on the CI
	// machine) hands over the waiting packets at once. The test's spurts are
	// longer apart, so a missing terminator between them still shows.
	maxStall = time.Second
)

// check splits one speaker's packets into spurts: a terminator or a pause
// ends one (not a stall, see maxStall). A last spurt without a terminator
// that was still running at end is open, not wrong.
func check(ps []packet, end time.Time) report {
	var r report
	var cur *spurt
	var prev packet
	for i, p := range ps {
		pause := p.at.Sub(prev.at)
		stall := cur != nil && !prev.terminator && pause > spurtPause && pause <= maxStall && p.frame == prev.frame+frameStep
		if stall {
			cur.stalls++
			cur.longestStall = max(cur.longestStall, pause)
		}
		if cur == nil || prev.terminator || pause > spurtPause && !stall {
			r.spurts = append(r.spurts, spurt{first: p.frame, afterTerminator: i > 0 && prev.terminator && p.at.Sub(prev.at) < extraGap})
			cur = &r.spurts[len(r.spurts)-1]
		} else if d := int64(p.frame) - int64(prev.frame); d <= 0 {
			cur.repeats++
		} else if d > frameStep {
			cur.gaps++
		}
		cur.packets++
		cur.last = p.frame
		if p.terminator {
			cur.terminators++
		}
		prev = p
	}
	if cur != nil && cur.terminators == 0 && end.Sub(prev.at) <= spurtPause {
		cur.open = true
	}
	return r
}

func (s spurt) ok() bool {
	return (s.terminators == 1 || s.open && s.terminators == 0) && s.gaps == 0 && s.repeats == 0 && !s.afterTerminator
}

func (r report) ok() bool {
	for _, s := range r.spurts {
		if !s.ok() {
			return false
		}
	}
	return len(r.spurts) > 0
}

func (s spurt) String() string {
	var bad []string
	if s.open {
		bad = append(bad, "still running")
	} else if s.terminators != 1 {
		bad = append(bad, fmt.Sprintf("%d terminators", s.terminators))
	}
	if s.gaps > 0 {
		bad = append(bad, fmt.Sprintf("%d gaps", s.gaps))
	}
	if s.repeats > 0 {
		bad = append(bad, fmt.Sprintf("%d repeats", s.repeats))
	}
	if s.afterTerminator {
		bad = append(bad, "extra packet after the terminator")
	}
	status := "ok"
	if s.open && len(bad) == 1 {
		status = "ok, still running"
	} else if len(bad) > 0 {
		status = strings.Join(bad, ", ")
	}
	if s.stalls > 0 {
		word := "stalls"
		if s.stalls == 1 {
			word = "stall"
		}
		status += fmt.Sprintf(", %d %s (longest %d ms)", s.stalls, word, s.longestStall.Milliseconds())
	}
	return fmt.Sprintf("frames %d-%d, %d packets: %s", s.first, s.last, s.packets, status)
}
