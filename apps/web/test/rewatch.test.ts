import { describe, expect, it } from "vitest";
import { follow, NOTHING_WATCHED, rewatch, type Seen } from "../src/lib/stream/rewatch";

describe("rewatch", () => {
  const lost = [{ name: "Lain", camera: false, until: 1000 }];
  it("watches the same user's stream again when it comes back", () => {
    expect(rewatch(lost, [{ id: 7, name: "lain", camera: false }], 500)).toEqual({ watch: [7], lost: [] });
  });
  it("not a camera for a screen, nor someone else's", () => {
    expect(rewatch(lost, [{ id: 7, name: "Lain", camera: true }, { id: 8, name: "sam", camera: false }], 500)).toEqual({ watch: [], lost });
  });
  it("forgets after a while", () => {
    expect(rewatch(lost, [], 1000)).toEqual({ watch: [], lost: [] });
    expect(rewatch(lost, [{ id: 7, name: "Lain", camera: false }], 1001)).toEqual({ watch: [], lost: [] });
  });
});

describe("follow", () => {
  const lain = (id: number) => ({ id, name: "lain", camera: false, known: true });
  const seen = (o: Partial<Seen>): Seen => ({ here: [], live: new Set(), watching: [], closed: new Set(), inVoice: true, now: 0, ...o });
  // Watching stream 4, then its session leaves (the stream ends), then 7 comes.
  const run = (closed: ReadonlySet<number>) => {
    let w = follow(NOTHING_WATCHED, seen({ here: [lain(4)], live: new Set([4]), watching: [4] })).watched;
    // Closed by the viewer, or ended by the server: the stream client drops it either way.
    w = follow(w, seen({ here: [lain(4)], live: new Set([4]), watching: [], closed, now: 1 })).watched;
    w = follow(w, seen({ now: 2 })).watched;
    return follow(w, seen({ here: [lain(7)], live: new Set([7]), now: 3 }));
  };
  it("watches the same user's stream again after it ended", () => {
    expect(run(new Set()).watch).toEqual([7]);
  });
  it("not one the viewer closed", () => {
    expect(run(new Set([4])).watch).toEqual([]);
  });
  it("stops watching streams not here", () => {
    expect(follow(NOTHING_WATCHED, seen({ watching: [4], live: new Set([4]) })).unwatch).toEqual([4]);
  });
  it("keeps the first known name (the user list can lose it first)", () => {
    let w = follow(NOTHING_WATCHED, seen({ here: [lain(4)], live: new Set([4]), watching: [4] })).watched;
    w = follow(w, seen({ here: [{ id: 4, name: "someone", camera: false, known: false }], live: new Set([4]), watching: [4], now: 1 })).watched;
    w = follow(w, seen({ now: 2 })).watched;
    expect(follow(w, seen({ here: [lain(7)], live: new Set([7]), now: 3 })).watch).toEqual([7]);
  });
});
