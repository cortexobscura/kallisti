import { describe, expect, it } from "vitest";
import { dragShare, keyShare, parseShare, shareOf, STAGE_MAX, STAGE_MIN } from "../src/lib/stream/stageSize";

describe("stream view height", () => {
  it("starts at the default for the device until the user drags it", () => {
    expect(shareOf(null, false)).toBe(0.55);
    expect(shareOf(null, true)).toBe(0.4);
    expect(shareOf(0.3, true)).toBe(0.3);
  });

  it("reads a saved share, and nothing from broken data", () => {
    expect(parseShare("0.35")).toBe(0.35);
    expect(parseShare("0.95")).toBe(STAGE_MAX);
    expect(parseShare("0.01")).toBe(STAGE_MIN);
    expect(parseShare(null)).toBeNull();
    expect(parseShare('"tall"')).toBeNull();
    expect(parseShare("{")).toBeNull();
  });

  it("follows the pointer, within the limits", () => {
    expect(dragShare(0.5, 400, 500, 1000)).toBeCloseTo(0.6);
    expect(dragShare(0.5, 400, 300, 1000)).toBeCloseTo(0.4);
    expect(dragShare(0.5, 400, 1000, 1000)).toBe(STAGE_MAX);
    expect(dragShare(0.5, 400, 0, 1000)).toBe(STAGE_MIN);
  });

  it("moves in steps with the arrow keys, to the limits with Home and End", () => {
    expect(keyShare(0.5, "ArrowDown")).toBeCloseTo(0.55);
    expect(keyShare(0.5, "ArrowUp")).toBeCloseTo(0.45);
    expect(keyShare(0.5, "Home")).toBe(STAGE_MIN);
    expect(keyShare(0.5, "End")).toBe(STAGE_MAX);
    expect(keyShare(STAGE_MAX, "ArrowDown")).toBe(STAGE_MAX);
    expect(keyShare(0.5, "Enter")).toBeNull();
  });

  it("keeps three digits (no float rest in storage)", () => {
    expect(keyShare(0.3, "ArrowDown")).toBe(0.35);
    expect(dragShare(0.5, 400, 300.12345, 1000)).toBe(0.4);
  });
});
