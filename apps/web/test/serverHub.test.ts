// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { createServerHub } from "../src/lib/serverHub.svelte";
import { loadTokens, type SavedServer, saveTokens } from "../src/lib/serverList";

const memory = () => {
  const m = new Map<string, string>();
  return { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => void m.set(k, v), removeItem: (k: string) => void m.delete(k) } as Storage;
};
const HOME = "https://chat.example";
const A = "https://a.example";
const B = "https://b.example";
const flush = () => new Promise((r) => setTimeout(r, 0));

// The home account's list (on the server), shared by the hubs of a test.
const account = (initial: SavedServer[] = []) => {
  let list = [...initial];
  const api = {
    fail: false,
    load: vi.fn(async () => [...list]),
    add: vi.fn(async (s: SavedServer) => {
      if (api.fail) throw new Error("no answer");
      list = list.some((x) => x.origin === s.origin) ? list.map((x) => (x.origin === s.origin ? s : x)) : [...list, s];
    }),
    remove: vi.fn(async (origin: string) => void (list = list.filter((x) => x.origin !== origin))),
    get list() {
      return list;
    },
  };
  return api;
};

const setup = (api = account(), storage = memory()) => {
  let owner: string | null = "home alice";
  const revoked: [string, string][] = [];
  const sessions: { origin: string; name: string; token: string | null; leave: ReturnType<typeof vi.fn>; dispose: ReturnType<typeof vi.fn>; close: ReturnType<typeof vi.fn> }[] = [];
  const hub = createServerHub({
    storage: () => storage,
    allowed: (origin) => origin !== HOME,
    list: () => (owner === null ? null : api),
    owner: () => owner,
    open: (origin, name, token) => {
      const s = { origin, name, token, leave: vi.fn(), dispose: vi.fn(), close: vi.fn() };
      sessions.push(s);
      return s;
    },
    revoke: async (origin, token) => void revoked.push([origin, token]),
  });
  hub.follow(owner);
  return { hub, api, storage, revoked, sessions, login: (o: string | null) => void (owner = o) };
};

describe("the added servers", () => {
  it("come from the home account; the ones with a token here connect, the others ask for a login", async () => {
    const storage = memory();
    saveTokens(storage, "home alice", { [A]: { token: "ta", name: "alice" } });
    const { hub, sessions } = setup(account([{ origin: A, name: "alice" }, { origin: B, name: "al" }, { origin: HOME, name: "x" }]), storage);
    await hub.sync();
    expect(hub.servers.map((s) => [s.origin, s.name, s.token, s.session !== null])).toEqual([
      [A, "alice", "ta", true],
      [B, "al", null, false],
    ]);
    expect(sessions.map((s) => [s.origin, s.token])).toEqual([[A, "ta"]]);
  });

  it("are added: connected, shown, the token kept here, the list at home", async () => {
    const { hub, api, storage } = setup();
    expect(hub.add(A, "alice", "ta")).toBe(true);
    expect(hub.add(A, "bob", "tb")).toBe(false);
    expect(hub.add(HOME, "x", "th")).toBe(false);
    expect(hub.shown).toBe(A);
    expect(loadTokens(storage)).toEqual({ owner: "home alice", tokens: { [A]: { token: "ta", name: "alice", unsaved: true } } });
    await flush();
    expect(api.list).toEqual([{ origin: A, name: "alice" }]);
    expect(loadTokens(storage).tokens).toEqual({ [A]: { token: "ta", name: "alice" } }); // saved at home
  });

  it("need an account at home: guests add none", () => {
    const { hub, login } = setup();
    login(null);
    hub.follow(null);
    expect(hub.add(A, "alice", "ta")).toBe(false);
  });

  it("an add the home server does not save is undone, with an error", async () => {
    const api = account();
    api.fail = true;
    const { hub, revoked, sessions, storage } = setup(api);
    hub.add(A, "alice", "ta");
    await flush();
    expect(hub.servers).toEqual([]);
    expect(hub.shown).toBeNull();
    expect(sessions[0]!.leave).toHaveBeenCalled();
    expect(revoked).toEqual([[A, "ta"]]);
    expect(loadTokens(storage).tokens).toEqual({});
    expect(hub.error).toContain("no answer");
  });

  it("removing takes it out at home, logs out there and forgets the token", async () => {
    const { hub, api, revoked, sessions, storage } = setup();
    hub.add(A, "alice", "ta");
    await flush();
    hub.remove(A);
    await flush();
    expect(api.list).toEqual([]);
    expect(sessions[0]!.dispose).toHaveBeenCalled();
    expect(revoked).toEqual([[A, "ta"]]);
    expect(loadTokens(storage).tokens).toEqual({});
    expect(hub.servers).toEqual([]);
  });

  it("removed on another device: ended here at the next sync (an add still on its way stays)", async () => {
    const api = account();
    const { hub, revoked } = setup(api);
    hub.add(A, "alice", "ta");
    await flush();
    await api.remove(A); // another device
    api.add.mockImplementationOnce(() => new Promise(() => {})); // B's add never answers
    hub.add(B, "al", "tb");
    await hub.sync();
    expect(hub.servers.map((s) => s.origin)).toEqual([B]);
    expect(revoked).toEqual([[A, "ta"]]);
  });

  it("a server without a token here gets one by a login; a new name goes home", async () => {
    const api = account([{ origin: A, name: "alice" }]);
    const { hub, storage } = setup(api);
    await hub.sync();
    expect(hub.login(A, "alice2", "ta")).toBe(true);
    expect(hub.servers[0]!.session).not.toBeNull();
    expect(loadTokens(storage).tokens).toEqual({ [A]: { token: "ta", name: "alice2" } });
    await flush();
    expect(api.list).toEqual([{ origin: A, name: "alice2" }]);
  });

  it("a home logout ends them all: logged out there, tokens gone; the list stays at home", async () => {
    const { hub, api, revoked, sessions, storage } = setup();
    hub.add(A, "alice", "ta");
    await flush();
    hub.endAll();
    expect(sessions[0]!.leave).toHaveBeenCalled();
    expect(revoked).toEqual([[A, "ta"]]);
    expect(loadTokens(storage).tokens).toEqual({});
    expect(hub.servers).toEqual([]);
    expect(api.list).toEqual([{ origin: A, name: "alice" }]);
  });

  it("another home login ends the previous one's; its list loads anew", async () => {
    const { hub, revoked, login } = setup();
    hub.add(A, "alice", "ta");
    await flush();
    login("home bob");
    hub.follow("home bob");
    expect(hub.servers).toEqual([]);
    expect(revoked).toEqual([[A, "ta"]]);
  });

  it("tokens of another login on this device end at the start (a login that ended while no tab was open)", () => {
    const storage = memory();
    saveTokens(storage, "home bob", { [A]: { token: "tb", name: "bob" } });
    const { revoked } = setup(account(), storage);
    expect(revoked).toEqual([[A, "tb"]]);
    expect(loadTokens(storage)).toEqual({ owner: "home alice", tokens: {} });
  });

  it("a new page of the same login keeps the tokens", async () => {
    const storage = memory();
    saveTokens(storage, "home alice", { [A]: { token: "ta", name: "alice" } });
    const { revoked } = setup(account(), storage);
    expect(revoked).toEqual([]);
    expect(loadTokens(storage).tokens).toEqual({ [A]: { token: "ta", name: "alice" } });
  });

  it("a list that loads after another login counts for nothing", async () => {
    const api = account([{ origin: A, name: "alice" }]);
    let answer: (l: SavedServer[]) => void = () => {};
    api.load.mockImplementationOnce(() => new Promise((r) => (answer = r)));
    const { hub, login, sessions } = setup(api);
    const pending = hub.sync();
    login("home bob");
    hub.follow("home bob");
    answer([{ origin: A, name: "alice" }]);
    await pending;
    expect(hub.servers).toEqual([]);
    expect(sessions).toEqual([]);
  });

  it("a stale tab's logout does not end a newer login's tokens (another tab logged in since)", async () => {
    const storage = memory();
    const stale = setup(account(), storage);
    const fresh = setup(account(), storage); // another tab, same browser
    fresh.login("home bob");
    fresh.hub.follow("home bob");
    fresh.hub.add(A, "bob", "tb");
    await flush();
    stale.hub.endAll();
    expect(stale.revoked).toEqual([]);
    expect(loadTokens(storage)).toEqual({ owner: "home bob", tokens: { [A]: { token: "tb", name: "bob" } } });
  });

  it("a stale tab adds nothing over a newer login's tokens", () => {
    const storage = memory();
    const stale = setup(account(), storage);
    saveTokens(storage, "home bob", { [A]: { token: "tb", name: "bob" } }); // another tab logged in as bob
    expect(stale.hub.add(B, "alice", "ta")).toBe(false);
    expect(loadTokens(storage)).toEqual({ owner: "home bob", tokens: { [A]: { token: "tb", name: "bob" } } });
  });

  it("detach (another login in this browser) closes the open ones and keeps the tokens", async () => {
    const { hub, revoked, sessions, storage } = setup();
    hub.add(A, "alice", "ta");
    await flush();
    hub.detach();
    expect(sessions[0]!.leave).toHaveBeenCalled();
    expect(hub.servers).toEqual([]);
    expect(revoked).toEqual([]);
    expect(loadTokens(storage).tokens).toEqual({ [A]: { token: "ta", name: "alice" } });
  });

  it("a list loaded before an add does not drop it", async () => {
    const api = account();
    let answer: (l: SavedServer[]) => void = () => {};
    api.load.mockImplementationOnce(() => new Promise((r) => (answer = r)));
    const { hub, revoked } = setup(api);
    const pending = hub.sync();
    hub.add(A, "alice", "ta");
    await flush(); // its POST is done
    answer([]);
    await pending;
    expect(hub.servers.map((s) => s.origin)).toEqual([A]);
    expect(revoked).toEqual([]);
  });

  it("tokens stay on the device for no one: a page without a login ends them", () => {
    const storage = memory();
    saveTokens(storage, "home alice", { [A]: { token: "ta", name: "alice" } });
    const revoked: [string, string][] = [];
    const hub = createServerHub({
      storage: () => storage,
      allowed: (origin) => origin !== HOME,
      list: () => null,
      owner: () => null,
      open: () => ({ leave: () => {}, dispose: () => {}, close: () => {} }),
      revoke: async (origin, token) => void revoked.push([origin, token]),
    });
    hub.follow(null);
    expect(revoked).toEqual([[A, "ta"]]);
    expect(loadTokens(storage).tokens).toEqual({});
  });

  it("connect as the name of their token here (the list may name another device's login)", async () => {
    const storage = memory();
    saveTokens(storage, "home alice", { [A]: { token: "ta", name: "alice" } });
    const { hub, sessions } = setup(account([{ origin: A, name: "bob" }]), storage);
    await hub.sync();
    expect(sessions.map((s) => s.name)).toEqual(["alice"]);
    expect(hub.servers[0]!.name).toBe("alice");
  });

  it("only servers this client may reach come from the list", async () => {
    const { hub } = setup(account([{ origin: HOME, name: "x" }, { origin: A, name: "alice" }]));
    await hub.sync();
    expect(hub.servers.map((s) => s.origin)).toEqual([A]);
  });

  it("removed while this device was closed: its token here ends at the sync", async () => {
    const storage = memory();
    saveTokens(storage, "home alice", { [A]: { token: "ta", name: "alice" } });
    const { hub, revoked } = setup(account([]), storage);
    await hub.sync();
    expect(revoked).toEqual([[A, "ta"]]);
    expect(loadTokens(storage).tokens).toEqual({});
  });

  it("a login in another tab connects this tab's waiting entry at the next sync", async () => {
    const storage = memory();
    const api = account([{ origin: A, name: "alice" }]);
    const { hub, sessions } = setup(api, storage);
    await hub.sync();
    expect(hub.servers[0]!.session).toBeNull();
    saveTokens(storage, "home alice", { [A]: { token: "ta", name: "alice" } }); // the other tab
    await hub.sync();
    expect(hub.servers[0]!.token).toBe("ta");
    expect(sessions.map((s) => s.token)).toEqual(["ta"]);
  });

  it("another tab's add that is not saved yet keeps its token at this tab's sync", async () => {
    const storage = memory();
    const api = account([]);
    let answer: (l: SavedServer[]) => void = () => {};
    api.load.mockImplementationOnce(() => new Promise((r) => (answer = r)));
    const { hub, revoked } = setup(api, storage);
    const pending = hub.sync();
    saveTokens(storage, "home alice", { [A]: { token: "ta", name: "alice", unsaved: true } }); // the other tab adds
    answer([]);
    await pending;
    expect(revoked).toEqual([]);
    expect(loadTokens(storage).tokens).toEqual({ [A]: { token: "ta", name: "alice", unsaved: true } });
  });

  it("a list loaded while an add was on its way home does not drop it", async () => {
    const api = account();
    let answer: (l: SavedServer[]) => void = () => {};
    const { hub, revoked } = setup(api);
    hub.add(A, "alice", "ta");
    api.load.mockImplementationOnce(() => new Promise((r) => (answer = r)));
    const pending = hub.sync(); // its list is from before the add was saved
    await flush(); // the add is saved now
    answer([]);
    await pending;
    expect(hub.servers.map((s) => s.origin)).toEqual([A]);
    expect(revoked).toEqual([]);
  });

  it("tokens another tab stored while this page started are not ended at its first follow", () => {
    const storage = memory();
    const revoked: [string, string][] = [];
    const hub = createServerHub({
      storage: () => storage,
      allowed: () => true,
      list: () => null,
      owner: () => null,
      open: () => ({ leave: () => {}, dispose: () => {}, close: () => {} }),
      revoke: async (origin, token) => void revoked.push([origin, token]),
    });
    saveTokens(storage, "home bob", { [A]: { token: "tb", name: "bob" } }); // the other tab
    hub.follow(null); // this page's late first answer: not logged in
    expect(revoked).toEqual([]);
  });

  it("loaded only when the list came (a failed load is listed, not loaded)", async () => {
    const api = account([{ origin: A, name: "alice" }]);
    api.load.mockImplementationOnce(async () => {
      throw new Error("down");
    });
    const { hub } = setup(api);
    await hub.sync();
    expect([hub.listed, hub.loaded]).toEqual([true, false]);
    await hub.sync();
    expect([hub.listed, hub.loaded]).toEqual([true, true]);
  });
});
