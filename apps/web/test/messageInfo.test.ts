import { describe, expect, it } from "vitest";
import { infoRows, longDate } from "../src/lib/messageInfo";

const at = Date.UTC(2026, 9, 26, 15, 58);

describe("message info", () => {
  it("writes a full date and time in the chosen clock", () => {
    expect(longDate(at, "en-US", "24")).toMatch(/^October 26, 2026 at \d\d:58$/);
    expect(longDate(at, "en-US", "12")).toMatch(/^October 26, 2026 at \d{1,2}:58 (AM|PM)$/);
  });
  it("has sent, and edited and pinned when they happened", () => {
    expect(infoRows({ at }, "en-US", "24").map((r) => r.label)).toEqual(["Sent"]);
    const rows = infoRows({ at, edited: at + 60_000, pinned: at + 120_000 }, "en-US", "24");
    expect(rows.map((r) => r.label)).toEqual(["Sent", "Edited", "Pinned"]);
    expect(rows[1]!.value).toBe(longDate(at + 60_000, "en-US", "24"));
  });
});
