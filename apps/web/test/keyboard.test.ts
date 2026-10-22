import { describe, expect, it } from "vitest";
import { keyboardOpen, watchKeyboard } from "../src/lib/keyboard";

describe("keyboardOpen", () => {
  it("is open when the visual viewport is much shorter than the layout", () => {
    expect(keyboardOpen({ layout: 812, visual: 471, scale: 1 })).toBe(true);
  });
  it("is closed when they are about the same", () => {
    expect(keyboardOpen({ layout: 812, visual: 812, scale: 1 })).toBe(false);
    expect(keyboardOpen({ layout: 812, visual: 760, scale: 1 })).toBe(false);
  });
  it("does not take pinch zoom for a keyboard", () => {
    expect(keyboardOpen({ layout: 812, visual: 406, scale: 2 })).toBe(false);
  });
});

const fakeViewport = (height: number) => {
  const listeners = new Set<() => void>();
  return {
    height,
    scale: 1,
    addEventListener: (_: string, cb: () => void) => void listeners.add(cb),
    removeEventListener: (_: string, cb: () => void) => void listeners.delete(cb),
    resize(h: number) {
      this.height = h;
      listeners.forEach((cb) => cb());
    },
    listeners,
  };
};

describe("watchKeyboard", () => {
  it("tells each change, and nothing when it stays the same", () => {
    const vv = fakeViewport(812);
    const seen: boolean[] = [];
    watchKeyboard((open) => seen.push(open), { viewport: vv, layoutHeight: () => 812 });
    vv.resize(800);
    vv.resize(471);
    vv.resize(470);
    vv.resize(812);
    expect(seen).toEqual([true, false]);
  });
  it("stops listening", () => {
    const vv = fakeViewport(812);
    const stop = watchKeyboard(() => {}, { viewport: vv, layoutHeight: () => 812 });
    stop();
    expect(vv.listeners.size).toBe(0);
  });
});
