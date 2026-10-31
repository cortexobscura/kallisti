// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { createFriends } from "../src/lib/friends.svelte";
import type { Friend } from "../src/lib/friends";

const bob: Friend = { origin: "https://b.example", user: 7, name: "bob" };
const carol: Friend = { origin: "https://a.example", user: 3, name: "carol" };

const device = (start: Friend[] = []) => {
  let saved = start;
  return { get: () => saved, set: vi.fn((list: Friend[]) => void (saved = list)), saved: () => saved };
};
// A promise the test settles.
const later = <T>() => {
  let resolve!: (v: T) => void;
  let reject!: (e: Error) => void;
  const promise = new Promise<T>((res, rej) => ((resolve = res), (reject = rej)));
  return { promise, resolve, reject };
};
const api = (load: () => Promise<Friend[]>) => ({ load: vi.fn(load), add: vi.fn(async (_f: Friend) => {}), remove: vi.fn(async (_o: string, _u: number) => {}) });

describe("the friends state", () => {
  it("without a home account: on the device", async () => {
    const d = device([carol]);
    const f = createFriends({ owner: () => null, remote: () => api(async () => []), device: d });
    await f.sync();
    expect(f.list).toEqual([carol]);
    f.add(bob);
    expect(d.saved()).toEqual([carol, bob]);
    f.remove(carol.origin, carol.user);
    expect(d.saved()).toEqual([bob]);
  });

  it("with one: loaded from the home server, changes sent there (shown at once)", async () => {
    const remote = api(async () => [carol]);
    const d = device();
    const f = createFriends({ owner: () => "alice", remote: () => remote, device: d });
    await f.sync();
    expect(f.list).toEqual([carol]);
    f.add(bob);
    expect(f.list).toEqual([carol, bob]);
    await vi.waitFor(() => expect(remote.add).toHaveBeenCalledWith(bob));
    expect(d.set).not.toHaveBeenCalled();
  });

  it("another login: the list empties at once, and the old one's load does not come back", async () => {
    let owner = "alice";
    const aliceLoad = later<Friend[]>();
    const bobLoad = later<Friend[]>();
    const f = createFriends({ owner: () => owner, remote: (o) => api(() => (o === "alice" ? aliceLoad.promise : bobLoad.promise)), device: device() });
    void f.sync();
    owner = "bob";
    void f.sync();
    expect(f.list).toEqual([]);
    aliceLoad.resolve([carol]);
    await Promise.resolve();
    expect(f.list).toEqual([]);
    bobLoad.resolve([bob]);
    await vi.waitFor(() => expect(f.list).toEqual([bob]));
  });

  it("a change from an earlier login's dialog is not taken", async () => {
    let owner: string | null = "alice";
    const remote = api(async () => [carol]);
    const d = device();
    const f = createFriends({ owner: () => owner, remote: () => remote, device: d });
    await f.sync();
    owner = null; // logged out elsewhere, no sync yet
    f.remove(carol.origin, carol.user);
    expect(d.set).not.toHaveBeenCalled();
    expect(remote.remove).not.toHaveBeenCalled();
  });

  it("a change ends a load still running (its answer is older)", async () => {
    const remote = api(async () => [carol, bob]);
    const f = createFriends({ owner: () => "alice", remote: () => remote, device: device() });
    await f.sync();
    const stale = later<Friend[]>();
    remote.load.mockImplementationOnce(() => stale.promise);
    void f.sync();
    f.remove(bob.origin, bob.user);
    stale.resolve([carol, bob]);
    await Promise.resolve();
    expect(f.list).toEqual([carol]);
  });

  it("writes go one after the other", async () => {
    const remote = api(async () => []);
    const first = later<void>();
    remote.add.mockImplementationOnce(() => first.promise);
    const f = createFriends({ owner: () => "alice", remote: () => remote, device: device() });
    await f.sync();
    f.add(bob);
    f.remove(bob.origin, bob.user);
    await Promise.resolve();
    expect(remote.remove).not.toHaveBeenCalled();
    first.resolve();
    await vi.waitFor(() => expect(remote.remove).toHaveBeenCalled());
  });

  it("a refused write: told, and the server's list again", async () => {
    const remote = api(async () => [carol]);
    remote.add.mockImplementationOnce(async () => Promise.reject(new Error("full")));
    const onError = vi.fn();
    const f = createFriends({ owner: () => "alice", remote: () => remote, device: device(), onError });
    await f.sync();
    f.add(bob);
    await vi.waitFor(() => expect(onError).toHaveBeenCalled());
    await vi.waitFor(() => expect(f.list).toEqual([carol]));
  });
});
