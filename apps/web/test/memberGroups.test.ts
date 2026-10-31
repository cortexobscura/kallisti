import { describe, expect, it } from "vitest";
import { parseCollapsedGroups, toggleGroup } from "../src/lib/memberGroups";

describe("collapsed member groups", () => {
  it("toggles a group by its key", () => {
    const one = toggleGroup([], "offline");
    expect(one).toEqual(["offline"]);
    expect(toggleGroup(one, "friends")).toEqual(["offline", "friends"]);
    expect(toggleGroup(one, "offline")).toEqual([]);
  });

  it("reads the device's list (broken: none)", () => {
    expect(parseCollapsedGroups(JSON.stringify(["offline", 3, "friends"]))).toEqual(["offline", "friends"]);
    expect(parseCollapsedGroups("{")).toEqual([]);
    expect(parseCollapsedGroups(null)).toEqual([]);
  });
});
