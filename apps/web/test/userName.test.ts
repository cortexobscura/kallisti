import { describe, expect, it } from "vitest";
import { splitName } from "../src/lib/userName";

describe("a name and its home server", () => {
  it("splits an account from another server, as the fediverse writes it", () => {
    expect(splitName("alice@chat.example.org")).toEqual({ own: "alice", host: "chat.example.org" });
    expect(splitName("x.y@127.0.0.1:18799")).toEqual({ own: "x.y", host: "127.0.0.1:18799" });
    expect(splitName("alice")).toEqual({ own: "alice", host: null });
    // Not a host after the @ (older names of here): all one name.
    expect(splitName("a@b")).toEqual({ own: "a@b", host: null });
    expect(splitName("@x.org")).toEqual({ own: "@x.org", host: null });
  });
});
