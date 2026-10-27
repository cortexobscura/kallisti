import { describe, expect, it } from "vitest";
import { addServer, loadServers, removeServer, saveServers } from "../src/lib/serverList";

const memory = () => {
  const m = new Map<string, string>();
  return { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => void m.set(k, v), removeItem: (k: string) => void m.delete(k) } as Storage;
};

describe("the servers added on this device", () => {
  it("are kept as origins, in order, without duplicates or the home server", () => {
    const home = "https://chat.kallisti.chat";
    let list = addServer([], "https://test.kallisti.chat", home);
    list = addServer(list, "https://test.kallisti.chat", home);
    list = addServer(list, home, home);
    list = addServer(list, "https://other.example", home);
    expect(list).toEqual(["https://test.kallisti.chat", "https://other.example"]);
    expect(removeServer(list, "https://test.kallisti.chat")).toEqual(["https://other.example"]);
  });

  it("survive storage, and broken or foreign entries are dropped", () => {
    const s = memory();
    saveServers(s, ["https://a.example", "https://b.example"]);
    expect(loadServers(s)).toEqual(["https://a.example", "https://b.example"]);
    s.setItem("servers", "not json");
    expect(loadServers(s)).toEqual([]);
    s.setItem("servers", JSON.stringify(["https://ok.example", "javascript:alert(1)", 3, "https://x.example/path"]));
    expect(loadServers(s)).toEqual(["https://ok.example"]);
    expect(loadServers(undefined)).toEqual([]);
  });
});
