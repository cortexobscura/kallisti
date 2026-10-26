import { describe, expect, it } from "vitest";
import { createStreamStats, perSecond } from "../src/lib/stream/stats";

describe("stream stats", () => {
  it("counts pictures, sound, bytes and what went wrong", () => {
    const s = createStreamStats();
    s.picture(1000);
    s.picture(500);
    s.late("picture");
    s.drawn();
    s.skipped();
    for (const ts of [0, 20_000, 40_000, 100_000, 120_000, 140_000]) s.sound(80, ts); // 2 lost before 100 ms
    s.late("sound");
    expect(s.totals()).toEqual({ pictures: 2, sound: 6, bytes: 1980, drawn: 1, latePictures: 1, lateSound: 1, skipped: 1, lostSound: 2 });
  });
  it("does not count a new start as lost (a big jump or going back)", () => {
    const s = createStreamStats();
    for (const ts of [0, 20_000, 60_000_000, 10_000]) s.sound(80, ts);
    expect(s.totals().lostSound).toBe(0);
  });
  it("gives rates between two readings", () => {
    const a = { pictures: 10, sound: 50, bytes: 10_000, drawn: 9, latePictures: 0, lateSound: 0, skipped: 0, lostSound: 0 };
    const b = { ...a, pictures: 40, sound: 150, bytes: 260_000, drawn: 38 };
    expect(perSecond(a, b, 2)).toEqual({ pictures: 15, sound: 50, kbps: 1000, drawn: 14.5 });
  });
  it("gives no rates when the totals went back (a new viewer) or no time passed", () => {
    const a = { pictures: 10, sound: 50, bytes: 10_000, drawn: 9, latePictures: 0, lateSound: 0, skipped: 0, lostSound: 0 };
    expect(perSecond(a, { ...a, pictures: 2, drawn: 1 }, 1)).toBeNull();
    expect(perSecond(a, a, 0)).toBeNull();
  });
});
