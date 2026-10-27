// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { createServerHub } from "../src/lib/serverHub.svelte";
import { loadServers } from "../src/lib/serverList";

const memory = () => {
  const m = new Map<string, string>();
  return { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => void m.set(k, v), removeItem: (k: string) => void m.delete(k) } as Storage;
};
const HOME = "https://chat.example";
const A = "https://a.example";
const B = "https://b.example";

const setup = (storage = memory()) => {
  let owner: string | null = "home alice";
  const tokens = new Map<string, string | null>();
  const revoked: [string, string][] = [];
  const sessions: { origin: string; name: string; leave: ReturnType<typeof vi.fn>; dispose: ReturnType<typeof vi.fn>; close: ReturnType<typeof vi.fn> }[] = [];
  const hub = createServerHub({
    storage: () => storage,
    home: () => HOME,
    owner: () => owner,
    tokens: (origin) => ({ get: () => tokens.get(origin) ?? null, set: (t) => void tokens.set(origin, t) }),
    open: (origin, name) => {
      const s = { origin, name, leave: vi.fn(), dispose: vi.fn(), close: vi.fn() };
      sessions.push(s);
      return s;
    },
    revoke: async (origin, token) => void revoked.push([origin, token]),
  });
  return { hub, storage, tokens, revoked, sessions, login: (o: string | null) => void (owner = o) };
};

describe("the added servers", () => {
  it("are added once, kept with their token and name, and shown", () => {
    const { hub, storage, tokens } = setup();
    expect(hub.add(A, "alice", "ta")).toBe(true);
    expect(hub.add(A, "bob", "tb")).toBe(false); // the first login stays
    expect(hub.add(HOME, "x", "th")).toBe(false);
    expect(tokens.get(A)).toBe("ta");
    expect(hub.servers.map((s) => [s.origin, s.name])).toEqual([[A, "alice"]]);
    expect(hub.shown).toBe(A);
    expect(loadServers(storage, "home alice")).toEqual([{ origin: A, name: "alice" }]);
  });

  it("need an account at home: guests add none and restore none", () => {
    const { hub, login } = setup();
    login(null);
    expect(hub.add(A, "alice", "ta")).toBe(false);
    hub.restore();
    expect(hub.servers).toEqual([]);
  });

  it("are another home login's: restore ends them (logged out there, tokens and list gone)", () => {
    const first = setup();
    first.hub.add(A, "alice", "ta");
    const next = setup(first.storage);
    next.tokens.set(A, "ta"); // the same device
    next.login("home bob");
    next.hub.restore();
    expect(next.sessions).toEqual([]);
    expect(next.revoked).toEqual([[A, "ta"]]);
    expect(next.tokens.get(A)).toBeNull();
    expect(loadServers(first.storage)).toEqual([]);
  });

  it("removing forgets the token at once and logs out with the old one; a new login is not touched", async () => {
    const { hub, tokens, revoked, sessions, storage } = setup();
    hub.add(A, "alice", "ta");
    hub.remove(A);
    expect(tokens.get(A)).toBeNull();
    expect(sessions[0]!.leave).toHaveBeenCalled();
    expect(sessions[0]!.dispose).toHaveBeenCalled();
    expect(hub.shown).toBeNull();
    expect(loadServers(storage)).toEqual([]);
    hub.add(A, "alice", "ta2"); // again, before the logout answered
    await Promise.resolve();
    expect(revoked).toEqual([[A, "ta"]]);
    expect(tokens.get(A)).toBe("ta2");
  });

  it("restore opens the saved ones once per home login, with their names, at once", () => {
    const { hub, storage, sessions } = setup();
    hub.add(A, "alice", "ta");
    hub.add(B, "al", "tb");
    const again = setup(storage);
    again.hub.restore();
    again.hub.restore();
    expect(again.sessions.map((s) => [s.origin, s.name])).toEqual([[A, "alice"], [B, "al"]]);
    expect(again.hub.shown).toBeNull();
    expect(sessions).toHaveLength(2);
  });

  it("a home logout ends all of them: logged out there, tokens and list gone", () => {
    const { hub, tokens, revoked, sessions, storage } = setup();
    hub.add(A, "alice", "ta");
    hub.add(B, "al", "tb");
    hub.endAll();
    expect(hub.servers).toEqual([]);
    expect(hub.shown).toBeNull();
    expect(revoked).toEqual([[A, "ta"], [B, "tb"]]);
    expect([tokens.get(A), tokens.get(B)]).toEqual([null, null]);
    expect(loadServers(storage)).toEqual([]);
    expect(sessions.every((s) => s.leave.mock.calls.length === 1 && s.dispose.mock.calls.length === 1)).toBe(true);
    // The next login restores its own (none here).
    hub.restore();
    expect(hub.servers).toEqual([]);
  });

  it("close (the app goes away) closes the sessions and keeps list and tokens", () => {
    const { hub, tokens, sessions, storage, revoked } = setup();
    hub.add(A, "alice", "ta");
    hub.close();
    expect(sessions[0]!.close).toHaveBeenCalled();
    expect(tokens.get(A)).toBe("ta");
    expect(loadServers(storage)).toHaveLength(1);
    expect(revoked).toEqual([]);
  });

  it("show only shows servers in the list", () => {
    const { hub } = setup();
    hub.add(A, "alice", "ta");
    hub.show(null);
    hub.show(B);
    expect(hub.shown).toBeNull();
    hub.show(A);
    expect(hub.shown).toBe(A);
  });
});
