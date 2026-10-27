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
const C = "https://c.example";

// One clock for all hubs of a test (tabs of one browser).
let clock = 1000;
const setup = (storage = memory(), tokens = new Map<string, string | null>()) => {
  let owner: string | null = "home alice";
  const revoked: [string, string][] = [];
  const sessions: { origin: string; name: string; token: string | null; leave: ReturnType<typeof vi.fn>; dispose: ReturnType<typeof vi.fn>; close: ReturnType<typeof vi.fn> }[] = [];
  const hub = createServerHub({
    storage: () => storage,
    home: () => HOME,
    owner: () => owner,
    tokens: (origin) => ({ get: () => tokens.get(origin) ?? null, set: (t) => void tokens.set(origin, t) }),
    open: (origin, name, token) => {
      const s = { origin, name, token, leave: vi.fn(), dispose: vi.fn(), close: vi.fn() };
      sessions.push(s);
      return s;
    },
    revoke: async (origin, token) => void revoked.push([origin, token]),
    now: () => ++clock,
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

  it("follow: another home login, or none, ends the live ones at once (a cookie that expired in an open tab)", () => {
    const { hub, revoked, sessions, login, storage } = setup();
    hub.follow("home alice");
    hub.add(A, "alice", "ta");
    login(null);
    hub.follow(null);
    expect(hub.servers).toEqual([]);
    expect(revoked).toEqual([[A, "ta"]]);
    expect(sessions[0]!.leave).toHaveBeenCalled();
    expect(loadServers(storage)).toEqual([]);
    login("home bob");
    hub.follow("home bob");
    hub.restore();
    expect(hub.servers).toEqual([]); // bob restores his own (none)
  });

  it("follow at start: another login's saved ones end before anything opens (also for a guest)", () => {
    const first = setup();
    first.hub.add(A, "alice", "ta");
    const next = setup(first.storage);
    next.tokens.set(A, "ta");
    next.login(null);
    next.hub.follow(null);
    expect(next.revoked).toEqual([[A, "ta"]]);
    expect(loadServers(first.storage)).toEqual([]);
  });

  it("follow keeps the same login's (a reload, a reconnect)", () => {
    const { hub, revoked } = setup();
    hub.follow("home alice");
    hub.add(A, "alice", "ta");
    hub.follow("home alice");
    expect(hub.servers).toHaveLength(1);
    expect(revoked).toEqual([]);
  });

  it("an old list (an array, before owners) is ended, not restored", () => {
    const storage = memory();
    storage.setItem("servers", JSON.stringify([{ origin: A, name: "alice" }]));
    const { hub, tokens, revoked } = setup(storage);
    tokens.set(A, "ta");
    hub.follow("home alice");
    hub.restore();
    expect(hub.servers).toEqual([]);
    expect(revoked).toEqual([[A, "ta"]]);
  });

  it("a stale tab ending its login does not touch a newer login's list or tokens", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const stale = setup(storage, tokens);
    stale.hub.follow("home alice");
    stale.hub.add(A, "alice", "ta");
    // Another tab: alice out, bob in, bob adds A too.
    const fresh = setup(storage, tokens);
    fresh.login("home bob");
    fresh.hub.follow("home bob");
    fresh.hub.add(A, "bob", "tb");
    // The stale tab learns that alice is out.
    stale.login(null);
    stale.hub.follow(null);
    expect(stale.hub.servers).toEqual([]);
    expect(stale.sessions[0]!.leave).toHaveBeenCalled();
    expect(tokens.get(A)).toBe("tb");
    expect(stale.revoked).toEqual([[A, "ta"]]); // its own old token, not bob's
    expect(loadServers(storage, "home bob")).toEqual([{ origin: A, name: "bob" }]);
  });

  it("sessions keep the token they were opened with; restore then logout logs out the restored token", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const first = setup(storage, tokens);
    first.hub.add(A, "alice", "ta");
    expect(first.sessions[0]!.token).toBe("ta");
    const again = setup(storage, tokens);
    again.hub.follow("home alice");
    again.hub.restore();
    expect(again.sessions[0]!.token).toBe("ta");
    again.hub.endAll();
    expect(again.revoked).toEqual([[A, "ta"]]);
  });

  it("a stale tab does not add over another login's saved list", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const stale = setup(storage, tokens);
    stale.hub.follow("home alice");
    const fresh = setup(storage, tokens);
    fresh.login("home bob");
    fresh.hub.follow("home bob");
    fresh.hub.add(B, "bob", "tb");
    expect(stale.hub.add(A, "alice", "ta")).toBe(false);
    expect(tokens.get(A)).toBeUndefined();
    expect(loadServers(storage, "home bob")).toEqual([{ origin: B, name: "bob" }]);
  });

  it("a delayed first follow does not end a list another tab saved after this tab started", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const slow = setup(storage, tokens); // started, its check still under way
    const other = setup(storage, tokens);
    other.login("home bob");
    other.hub.follow("home bob");
    other.hub.add(B, "bob", "tb");
    slow.hub.follow("home alice"); // the old answer
    slow.hub.restore();
    expect(tokens.get(B)).toBe("tb");
    expect(loadServers(storage, "home bob")).toEqual([{ origin: B, name: "bob" }]);
  });

  it("a late logout of the same account does not sweep that account's newer list", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const stale = setup(storage, tokens);
    stale.hub.follow("home alice");
    stale.hub.add(A, "alice", "ta");
    const fresh = setup(storage, tokens);
    fresh.hub.follow("home alice");
    fresh.hub.restore();
    // There: logged out, and in again as alice (a new login: another session).
    fresh.login(null);
    fresh.hub.follow(null);
    fresh.login("home alice 2");
    fresh.hub.follow("home alice 2");
    fresh.hub.restore();
    fresh.hub.add(B, "alice", "tb2");
    stale.login(null);
    stale.hub.follow(null);
    expect(tokens.get(B)).toBe("tb2");
    expect(loadServers(storage, "home alice 2")).toEqual([{ origin: B, name: "alice" }]);
  });

  it("remove from a stale tab does not touch another login's token or list", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const stale = setup(storage, tokens);
    stale.hub.follow("home alice");
    stale.hub.add(A, "alice", "ta");
    const fresh = setup(storage, tokens);
    fresh.login("home bob");
    fresh.hub.follow("home bob");
    fresh.hub.add(A, "bob", "tb");
    stale.hub.remove(A);
    expect(tokens.get(A)).toBe("tb");
    expect(loadServers(storage, "home bob")).toEqual([{ origin: A, name: "bob" }]);
  });

  it("two tabs of one login share the list: a remove in one is kept, a logout in the other ends all", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const a = setup(storage, tokens);
    a.hub.follow("home alice");
    a.hub.add(A, "alice", "ta");
    const b = setup(storage, tokens);
    b.hub.follow("home alice");
    b.hub.restore();
    b.hub.add(B, "alice", "tb");
    a.hub.remove(A);
    expect(loadServers(storage, "home alice")).toEqual([{ origin: B, name: "alice" }]);
    a.login(null);
    a.hub.follow(null); // A's logout: B's server is this login's too
    expect(tokens.get(B)).toBeNull();
    expect(loadServers(storage)).toEqual([]);
  });

  it("a logout in a tab that never restored ends this login's older list", () => {
    const first = setup();
    first.hub.add(A, "alice", "ta");
    const later = setup(first.storage, first.tokens);
    later.hub.follow("home alice"); // its home never synced: no restore
    later.hub.endAll();
    expect(first.tokens.get(A)).toBeNull();
    expect(loadServers(first.storage)).toEqual([]);
  });

  it("ending a login logs out its own token there even if another tab stored a newer one", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const a = setup(storage, tokens);
    a.hub.follow("home alice");
    a.hub.add(A, "alice", "t1");
    tokens.set(A, "t2"); // another tab of this login logged in there again
    a.hub.remove(A);
    expect(a.revoked).toEqual([[A, "t1"]]);
    expect(tokens.get(A)).toBe("t2");
  });

  it("an old tab of the same account adds nothing to a newer login's list", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const old = setup(storage, tokens);
    old.hub.follow("home alice");
    old.hub.add(A, "alice", "ta");
    const fresh = setup(storage, tokens);
    fresh.hub.follow("home alice");
    fresh.hub.restore();
    fresh.login(null);
    fresh.hub.follow(null);
    fresh.login("home alice 2");
    fresh.hub.follow("home alice 2");
    fresh.hub.restore();
    fresh.hub.add(B, "alice", "tb");
    expect(old.hub.add(C, "alice", "tc")).toBe(false);
    expect(loadServers(storage, "home alice 2")).toEqual([{ origin: B, name: "alice" }]);
  });

  it("a tab that never restored ends this login's list at logout, also a newer one of this login", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const a = setup(storage, tokens);
    a.hub.follow("home alice"); // its home never synced
    const c = setup(storage, tokens);
    c.hub.follow("home alice");
    c.hub.add(A, "alice", "ta");
    a.hub.endAll(); // alice logs out in tab a
    expect(tokens.get(A)).toBeNull();
    expect(loadServers(storage)).toEqual([]);
  });

  it("after a login ends, the next login in that tab adds again", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const a = setup(storage, tokens);
    const b = setup(storage, tokens);
    a.hub.follow("home alice");
    a.hub.add(A, "alice", "ta");
    b.hub.follow("home alice");
    b.hub.restore();
    a.hub.endAll(); // logged out in a: b learns it too
    b.login(null);
    b.hub.follow(null);
    b.login("home alice 2"); // the new login (one cookie: both tabs)
    b.hub.follow("home alice 2");
    b.hub.restore();
    a.login("home alice 2");
    a.hub.follow("home alice 2");
    a.hub.restore();
    a.hub.add(A, "alice", "ta2");
    expect(b.hub.add(B, "alice", "tb")).toBe(true);
  });

  it("a server another tab of this login added cannot be added again (its token would be lost)", () => {
    const storage = memory();
    const tokens = new Map<string, string | null>();
    const a = setup(storage, tokens);
    const b = setup(storage, tokens);
    a.hub.follow("home alice");
    b.hub.follow("home alice");
    a.hub.add(A, "alice", "ta");
    expect(b.hub.add(A, "alice", "tb")).toBe(false);
    expect(tokens.get(A)).toBe("ta");
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
