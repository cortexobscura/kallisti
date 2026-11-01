import { describe, expect, it } from "vitest";
import { sectionsFor, startSection } from "../src/lib/settingsSections";

describe("settings sections", () => {
  it("shows the categories there is something for", () => {
    expect(sectionsFor({ profile: true, account: true }).map((s) => s.id)).toEqual(["profile", "account", "voice", "notifications", "appearance"]);
    expect(sectionsFor({ profile: false, account: false }).map((s) => s.id)).toEqual(["voice", "notifications", "appearance"]);
  });

  it("groups them as user and app settings", () => {
    expect(sectionsFor({ profile: true, account: true }).map((s) => s.group)).toEqual(["user", "user", "app", "app", "app"]);
  });

  it("starts at the asked one (if there), else: phones at the list, desktop at the first", () => {
    const all = sectionsFor({ profile: true, account: true });
    expect(startSection("account", all, true)).toBe("account");
    expect(startSection("account", all, false)).toBe("account");
    expect(startSection(undefined, all, true)).toBeNull();
    expect(startSection(undefined, all, false)).toBe("profile");
    const few = sectionsFor({ profile: false, account: false });
    expect(startSection("account", few, false)).toBe("voice");
  });
});
