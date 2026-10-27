import { describe, expect, it } from "vitest";
import { railBadge, tabTitle } from "../src/lib/title";

describe("tabTitle", () => {
  it("shows mentions and direct messages not read, as Discord", () => {
    expect(tabTitle("Kallisti", 0)).toBe("Kallisti");
    expect(tabTitle("Kallisti", 3)).toBe("(3) Kallisti");
    expect(tabTitle("My Server", 120)).toBe("(99+) My Server");
  });
});

describe("a server's mark in the rail", () => {
  it("counts mentions and direct messages; a dot for other unread text", () => {
    expect(railBadge({ 3: 2, 5: 1 }, { ann: 2 }, new Set([3]))).toEqual({ mentions: 5, unread: true });
    expect(railBadge({}, {}, new Set([4]))).toEqual({ mentions: 0, unread: true });
    expect(railBadge({}, {}, new Set())).toEqual({ mentions: 0, unread: false });
  });
});
