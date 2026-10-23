import { describe, expect, it } from "vitest";
import { rewatch } from "../src/lib/stream/rewatch";

describe("rewatch", () => {
  const lost = [{ name: "Lain", camera: false, until: 1000 }];
  it("watches the same user's stream again when it comes back", () => {
    expect(rewatch(lost, [{ id: 7, name: "lain", camera: false }], 500)).toEqual({ watch: [7], lost: [] });
  });
  it("not a camera for a screen, nor someone else's", () => {
    expect(rewatch(lost, [{ id: 7, name: "Lain", camera: true }, { id: 8, name: "sam", camera: false }], 500)).toEqual({ watch: [], lost });
  });
  it("forgets after a while", () => {
    expect(rewatch(lost, [], 1000)).toEqual({ watch: [], lost: [] });
    expect(rewatch(lost, [{ id: 7, name: "Lain", camera: false }], 1001)).toEqual({ watch: [], lost: [] });
  });
});
