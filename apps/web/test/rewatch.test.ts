import { describe, expect, it } from "vitest";
import { follow, leaving, loadLost, NOTHING_WATCHED, REWATCH_MS, rewatch, saveLost, type Seen } from "../src/lib/stream/rewatch";

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

describe("leaving", () => {
  it("makes the watched streams lost for a while, with those still waited for", () => {
    const waited = [{ name: "sam", camera: true, until: 500 }, { name: "old", camera: false, until: 50 }];
    expect(leaving([{ name: "Lain", camera: false }], waited, 100)).toEqual([
      { name: "sam", camera: true, until: 500 },
      { name: "Lain", camera: false, until: 100 + REWATCH_MS },
    ]);
  });
  it("keeps one entry for a stream watched and waited for", () => {
    expect(leaving([{ name: "lain", camera: false }], [{ name: "Lain", camera: false, until: 500 }], 100)).toEqual([
      { name: "lain", camera: false, until: 100 + REWATCH_MS },
    ]);
  });
});

describe("the memory over a reload", () => {
  const storage = () => {
    const m = new Map<string, string>();
    return { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => void m.set(k, v), removeItem: (k: string) => void m.delete(k) };
  };
  it("comes back while it is fresh", () => {
    const s = storage();
    saveLost(s, "k", [{ name: "lain", camera: false, until: 500 }, { name: "sam", camera: true, until: 90 }]);
    expect(loadLost(s, "k", 100)).toEqual([{ name: "lain", camera: false, until: 500 }]);
    expect(loadLost(s, "other", 100)).toEqual([]);
  });
  it("ignores broken data and missing storage", () => {
    const s = storage();
    s.setItem("k", "{nope");
    expect(loadLost(s, "k", 0)).toEqual([]);
    s.setItem("k", JSON.stringify([{ name: 1, camera: false, until: 5 }]));
    expect(loadLost(s, "k", 0)).toEqual([]);
    expect(loadLost(undefined, "k", 0)).toEqual([]);
    expect(() => saveLost(undefined, "k", [])).not.toThrow();
  });
});

describe("over a reconnect or a reload", () => {
  it("a new Client watches the stream again when it comes back in time", () => {
    const lain = (id: number) => ({ id, name: "lain", camera: false, known: true });
    const seen = (o: Partial<Seen>): Seen => ({ here: [], live: new Set(), watching: [], closed: new Set(), inVoice: true, now: 0, ...o });
    const before = follow(NOTHING_WATCHED, seen({ here: [lain(4)], live: new Set([4]), watching: [4] })).watched;
    const watching = [...before.names].map(([, name]) => ({ name, camera: false }));
    const lost = leaving(watching, before.lost, 1000);
    // The new Client: voice not back yet, then the stream under a new id.
    let w = follow({ ...NOTHING_WATCHED, lost }, seen({ inVoice: false, now: 2000 })).watched;
    expect(follow(w, seen({ here: [lain(9)], live: new Set([9]), now: 5000 })).watch).toEqual([9]);
    w = follow({ ...NOTHING_WATCHED, lost }, seen({ now: 2000 })).watched;
    expect(follow(w, seen({ here: [lain(9)], live: new Set([9]), now: 1000 + REWATCH_MS })).watch).toEqual([]);
  });
});

// A lost connection first closes the streaming socket: nothing is watched
// while the old stream list still says live. follow then forgets the names,
// so the Client keeps the memory from before (it remembers only while both
// connections are up).
it("forgets the names when the streaming socket drops them first", () => {
  const seen = (o: Partial<Seen>): Seen => ({ here: [], live: new Set(), watching: [], closed: new Set(), inVoice: true, now: 0, ...o });
  const lain = { id: 4, name: "lain", camera: false, known: true };
  const w = follow(NOTHING_WATCHED, seen({ here: [lain], live: new Set([4]), watching: [4] })).watched;
  const after = follow(w, seen({ here: [lain], live: new Set([4]), watching: [], now: 1 })).watched;
  expect([...after.names]).toEqual([]);
  expect(after.lost).toEqual([]);
});
