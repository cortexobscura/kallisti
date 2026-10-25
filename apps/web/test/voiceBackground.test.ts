import { describe, expect, it, vi } from "vitest";
import { keepRunningHidden } from "../src/lib/voice/background";

const fakeDoc = () => {
  const listeners = new Set<() => void>();
  return {
    visibilityState: "visible" as DocumentVisibilityState,
    addEventListener: (_: "visibilitychange", l: () => void) => void listeners.add(l),
    removeEventListener: (_: "visibilitychange", l: () => void) => void listeners.delete(l),
    hide() {
      this.visibilityState = "hidden";
      listeners.forEach((l) => l());
    },
    show() {
      this.visibilityState = "visible";
      listeners.forEach((l) => l());
    },
    listeners,
  };
};

const fakeCtx = (state: string) => {
  const listeners = new Set<() => void>();
  const ctx = {
    state,
    resume: vi.fn(async () => void (ctx.state = "running")),
    addEventListener: (_: "statechange", l: () => void) => void listeners.add(l),
    removeEventListener: (_: "statechange", l: () => void) => void listeners.delete(l),
    set(next: string) {
      ctx.state = next;
      listeners.forEach((l) => l());
    },
    listeners,
  };
  return ctx;
};

describe("keepRunningHidden", () => {
  it("resumes a context WebKit interrupted when the page went to the background", () => {
    const doc = fakeDoc();
    const ctx = fakeCtx("running");
    keepRunningHidden(doc, ctx);
    ctx.set("interrupted"); // comes just before the page counts as hidden
    expect(ctx.resume).not.toHaveBeenCalled();
    doc.hide();
    expect(ctx.resume).toHaveBeenCalledOnce();
  });

  it("resumes a context interrupted while the page is hidden", () => {
    const doc = fakeDoc();
    const ctx = fakeCtx("running");
    keepRunningHidden(doc, ctx);
    doc.hide();
    ctx.set("interrupted");
    expect(ctx.resume).toHaveBeenCalledOnce();
  });

  it("leaves running, suspended and closed contexts alone", () => {
    const doc = fakeDoc();
    for (const state of ["running", "suspended", "closed"]) {
      const ctx = fakeCtx(state);
      const stop = keepRunningHidden(doc, ctx);
      doc.hide();
      doc.show();
      expect(ctx.resume).not.toHaveBeenCalled();
      stop();
    }
  });

  it("stops listening when stopped", () => {
    const doc = fakeDoc();
    const ctx = fakeCtx("running");
    const stop = keepRunningHidden(doc, ctx);
    stop();
    expect(doc.listeners.size).toBe(0);
    expect(ctx.listeners.size).toBe(0);
    ctx.set("interrupted");
    doc.hide();
    expect(ctx.resume).not.toHaveBeenCalled();
  });
});
