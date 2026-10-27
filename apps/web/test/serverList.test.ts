import { describe, expect, it } from "vitest";
import { loadTokens, savedServers, saveTokens, serverOrigin } from "../src/lib/serverList";

const memory = () => {
  const m = new Map<string, string>();
  return { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => void m.set(k, v), removeItem: (k: string) => void m.delete(k) } as Storage;
};

describe("a list from the server", () => {
  it("keeps valid entries only", () => {
    const a = { origin: "https://test.kallisti.chat", name: "alice" };
    expect(savedServers([a, { origin: "javascript:alert(1)", name: "x" }, { origin: "https://x.example/path", name: "x" }, { origin: "https://y.example" }, "https://z.example", 3])).toEqual([a]);
    expect(savedServers({})).toEqual([]);
  });
});

describe("the added servers' tokens on this device", () => {
  it("are kept for their owner (the home login); broken entries are dropped", () => {
    const s = memory();
    const ta = { token: "ta", name: "al" };
    saveTokens(s, "home alice", { "https://a.example": ta });
    expect(loadTokens(s)).toEqual({ owner: "home alice", tokens: { "https://a.example": ta } });
    s.setItem("added-sessions", JSON.stringify({ owner: "o", tokens: { "https://a.example": ta, "https://b.example": "tb", "https://c.example": { token: "" }, "not an origin": ta } }));
    expect(loadTokens(s)).toEqual({ owner: "o", tokens: { "https://a.example": ta } });
    s.setItem("added-sessions", "not json");
    expect(loadTokens(s)).toEqual({ owner: "", tokens: {} });
    expect(loadTokens(undefined)).toEqual({ owner: "", tokens: {} });
  });
});

describe("an address for another server", () => {
  const home = "https://chat.kallisti.chat";
  it("becomes its canonical origin, port kept, path dropped", () => {
    expect(serverOrigin("Test.Kallisti.Chat", home)).toEqual({ origin: "https://test.kallisti.chat" });
    expect(serverOrigin(" test.kallisti.chat/channels/3 ", home)).toEqual({ origin: "https://test.kallisti.chat" });
    expect(serverOrigin("https://other.example:8443/", home)).toEqual({ origin: "https://other.example:8443" });
    expect(serverOrigin("http://127.0.0.1:18800", "http://localhost:18799")).toEqual({ origin: "http://127.0.0.1:18800" });
    expect(serverOrigin("other.example:8443", home)).toEqual({ origin: "https://other.example:8443" });
    expect(serverOrigin("localhost:18800", "https://chat.example")).toEqual({ origin: "https://localhost:18800" });
    expect(serverOrigin("other.example.", home)).toEqual({ origin: "https://other.example" }); // one spelling
  });
  it("refuses plain http elsewhere, other schemes, nonsense and the home server", () => {
    for (const bad of ["http://other.example", "javascript:alert(1)", "ftp://x.example", "", "not a host!", "chat.kallisti.chat", "https://CHAT.kallisti.chat/"]) {
      expect(serverOrigin(bad, home), bad).toHaveProperty("error");
    }
  });
  it("refuses another port of the home server's host (it would get the home cookie), also spelled with a dot", () => {
    for (const bad of ["chat.kallisti.chat:8443", "https://chat.kallisti.chat:8443", "chat.kallisti.chat.", "https://chat.kallisti.chat.:8443"]) {
      expect(serverOrigin(bad, home), bad).toHaveProperty("error");
    }
  });
  it("in the desktop app (tokens, no home cookie) only the home origin itself is refused", () => {
    expect(serverOrigin("localhost:18800", "http://localhost:5173", false)).toEqual({ origin: "https://localhost:18800" });
  });
});
