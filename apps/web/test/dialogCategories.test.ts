import { describe, expect, it } from "vitest";
import { groupsOf, startCategory } from "../src/lib/dialogCategories";

const list = [
  { id: "a", title: "A", group: "one" },
  { id: "b", title: "B", group: "two" },
  { id: "c", title: "C", group: "one" },
];

describe("categories", () => {
  it("start at the asked one (if there), else: phones at the list, desktop at the first", () => {
    expect(startCategory("b", list, true)).toBe("b");
    expect(startCategory("x", list, false)).toBe("a");
    expect(startCategory(undefined, list, true)).toBeNull();
    expect(startCategory(undefined, [], false)).toBeNull();
  });

  it("group in the given order of groups, empty ones left out", () => {
    expect(groupsOf(list, { one: "One", two: "Two", three: "Three" })).toEqual([
      { id: "one", title: "One", categories: [list[0], list[2]] },
      { id: "two", title: "Two", categories: [list[1]] },
    ]);
  });
});
