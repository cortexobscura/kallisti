import { describe, expect, it } from "vitest";
import { addServer, loadServers, removeServer, saveServers, serverOrigin } from "../src/lib/serverList";

const memory = () => {
  const m = new Map<string, string>();
  return { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => void m.set(k, v), removeItem: (k: string) => void m.delete(k) } as Storage;
};

describe("the servers added on this device", () => {
  const a = { origin: "https://test.kallisti.chat", name: "alice" };
  const b = { origin: "https://other.example", name: "al" };
  it("are kept with the name there, in order, without duplicates or the home server", () => {
    const home = "https://chat.kallisti.chat";
    let list = addServer([], a, home);
    list = addServer(list, { ...a, name: "bob" }, home);
    list = addServer(list, { origin: home, name: "x" }, home);
    list = addServer(list, b, home);
    expect(list).toEqual([a, b]);
    expect(removeServer(list, a.origin)).toEqual([b]);
  });

  it("survive storage, and broken or foreign entries are dropped", () => {
    const s = memory();
    saveServers(s, [a, b]);
    expect(loadServers(s)).toEqual([a, b]);
    s.setItem("servers", "not json");
    expect(loadServers(s)).toEqual([]);
    s.setItem("servers", JSON.stringify([a, { origin: "javascript:alert(1)", name: "x" }, { origin: "https://x.example/path", name: "x" }, { origin: "https://y.example" }, "https://z.example", 3]));
    expect(loadServers(s)).toEqual([a]);
    expect(loadServers(undefined)).toEqual([]);
  });
});

describe("an address for another server", () => {
  const home = "https://chat.kallisti.chat";
  it("becomes its canonical origin, port kept, path dropped", () => {
    expect(serverOrigin("Test.Kallisti.Chat", home)).toEqual({ origin: "https://test.kallisti.chat" });
    expect(serverOrigin(" test.kallisti.chat/channels/3 ", home)).toEqual({ origin: "https://test.kallisti.chat" });
    expect(serverOrigin("https://other.example:8443/", home)).toEqual({ origin: "https://other.example:8443" });
    expect(serverOrigin("http://127.0.0.1:18800", "http://localhost:18799")).toEqual({ origin: "http://127.0.0.1:18800" });
  });
  it("refuses plain http elsewhere, other schemes, nonsense and the home server", () => {
    for (const bad of ["http://other.example", "javascript:alert(1)", "ftp://x.example", "", "not a host!", "chat.kallisti.chat", "https://CHAT.kallisti.chat/"]) {
      expect(serverOrigin(bad, home), bad).toHaveProperty("error");
    }
  });
  it("refuses another port of the home server's host (it would get the home cookie)", () => {
    expect(serverOrigin("chat.kallisti.chat:8443", home)).toHaveProperty("error");
  });
});
