import { describe, expect, it } from "vitest";
import { localStore, safeGet, safeSet } from "../src/lib/storage";

const broken = {
  getItem: () => {
    throw new Error("SecurityError");
  },
  setItem: () => {
    throw new Error("QuotaExceededError");
  },
};

describe("safe storage", () => {
  it("reads and writes normally", () => {
    const data = new Map<string, string>();
    const store = { getItem: (k: string) => data.get(k) ?? null, setItem: (k: string, v: string) => void data.set(k, v) };
    expect(safeSet(store, "a", "1")).toBe(true);
    expect(safeGet(store, "a")).toBe("1");
  });

  it("survives a storage that throws", () => {
    expect(safeGet(broken, "a")).toBeNull();
    expect(safeSet(broken, "a", "1")).toBe(false);
  });

  it("survives a missing storage", () => {
    expect(safeGet(undefined, "a")).toBeNull();
    expect(safeSet(undefined, "a", "1")).toBe(false);
  });
});

describe("localStore", () => {
  it("is localStorage, or undefined where it is blocked", () => {
    const real = Object.getOwnPropertyDescriptor(globalThis, "localStorage");
    const fake = { getItem: () => null, setItem: () => {} };
    try {
      Object.defineProperty(globalThis, "localStorage", { configurable: true, get: () => fake });
      expect(localStore()).toBe(fake);
      Object.defineProperty(globalThis, "localStorage", { configurable: true, get: () => { throw new Error("SecurityError"); } });
      expect(localStore()).toBeUndefined();
    } finally {
      if (real) Object.defineProperty(globalThis, "localStorage", real);
      else delete (globalThis as { localStorage?: unknown }).localStorage;
    }
  });
});
