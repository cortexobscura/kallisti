import { initialState } from "@kallisti/protocol";
import { describe, expect, it } from "vitest";
import { initials } from "../src/lib/avatar";
import { connectionError, proxyUrl } from "../src/lib/status";

describe("connectionError", () => {
  it("is null while connecting or synced and open", () => {
    expect(connectionError(initialState, false)).toBeNull();
    expect(connectionError({ ...initialState, status: "synced" }, false)).toBeNull();
  });

  it("uses the server reason when there is one", () => {
    const s = { ...initialState, status: "rejected" as const, error: "Wrong password" };
    expect(connectionError(s, true)).toBe("Wrong password");
  });

  it("has fallbacks for empty reasons", () => {
    expect(connectionError({ ...initialState, status: "rejected" }, false)).toMatch(/rejected/);
    expect(connectionError({ ...initialState, status: "kicked", error: "" }, false)).toMatch(
      /removed/,
    );
  });

  it("gives the server's reason for a closed socket (a guest refused on another site)", () => {
    expect(connectionError(initialState, true, "join as a guest on the server's own site")).toBe("join as a guest on the server's own site");
  });

  it("keeps technical close reasons to itself", () => {
    expect(connectionError(initialState, true, "upstream error")).toMatch(/Lost/);
    expect(connectionError(initialState, true, "join as a guest at evil.example")).toMatch(/Lost/);
    // Only the desktop app's own bridge says so; a server cannot.
    expect(connectionError(initialState, true, "cannot reach voice.example:64738: refused", true)).toBe("cannot reach voice.example:64738: refused");
    expect(connectionError(initialState, true, "cannot reach voice: log in again at evil.example")).toMatch(/Lost/);
  });

  it("reports a closed socket", () => {
    expect(connectionError({ ...initialState, status: "synced" }, true)).toMatch(/Lost/);
  });
});

describe("proxyUrl", () => {
  it("uses the same host, under /ws, with wss on https pages", () => {
    expect(proxyUrl({ protocol: "https:", host: "chat.example" })).toBe("wss://chat.example/ws");
    expect(proxyUrl({ protocol: "http:", host: "localhost:8080" })).toBe("ws://localhost:8080/ws");
  });
});

describe("initials with emoji", () => {
  it("does not split surrogate pairs", () => {
    expect(initials("🎮gamer")).toBe("🎮");
  });
});
