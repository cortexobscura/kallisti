import { describe, expect, it } from "vitest";
import { befriendable, friendPresence, isFriend, parseFriends, withFriend, withoutFriend, type Friend, type FriendServer } from "../src/lib/friends";

const bob: Friend = { origin: "https://b.example", user: 7, name: "bob" };
const carol: Friend = { origin: "https://a.example", user: 3, name: "carol" };

describe("the friends list", () => {
  it("reads the server's list, dropping what is not a friend", () => {
    expect(parseFriends([bob, { origin: "x" }, { origin: "https://a.example", user: -1, name: "n" }, carol])).toEqual([bob, carol]);
    expect(parseFriends(null)).toEqual([]);
    expect(parseFriends([{ ...bob, origin: "tauri://localhost" }, { ...bob, name: "" }, { ...bob, origin: "https://" }])).toEqual([]);
    expect(parseFriends([bob, { ...bob, name: "again" }])).toEqual([bob]);
  });

  it("takes members only (no guest accounts, no users without an account)", () => {
    const members = { 7: "bob" };
    expect(befriendable(7, members)).toBe(true);
    expect(befriendable(8, members)).toBe(false); // a guest account
    expect(befriendable(null, members)).toBe(false);
  });

  it("adds (a known one: the new name, the same place) and removes by server and user id", () => {
    const list = withFriend(withFriend([], bob), carol);
    expect(withFriend(list, { ...bob, name: "bobby" })).toEqual([{ ...bob, name: "bobby" }, carol]);
    expect(isFriend(list, "https://b.example", 7)).toBe(true);
    expect(isFriend(list, "https://a.example", 7)).toBe(false);
    expect(withoutFriend(list, "https://b.example", 7)).toEqual([carol]);
  });
});

describe("friendPresence", () => {
  const servers: FriendServer[] = [
    {
      origin: "https://b.example",
      name: "Bob's place",
      online: { 7: 12 },
      users: { 12: { channelId: 5 } },
      voiceName: (id: number) => (id === 5 ? "Games" : null),
    },
    { origin: "https://c.example", name: "C", online: {}, users: {}, voiceName: () => null },
  ];

  it("finds friends online on the user's servers, with the voice channel they are in", () => {
    const rows = friendPresence([carol, bob], servers);
    expect(rows).toEqual([
      { friend: bob, server: "Bob's place", online: true, voice: "Games" },
      { friend: carol, server: null, online: false, voice: null },
    ]);
  });

  it("knows the server's name while offline, and no voice channel outside one", () => {
    const rows = friendPresence([{ origin: "https://c.example", user: 1, name: "d" }], servers);
    expect(rows[0]).toEqual({ friend: { origin: "https://c.example", user: 1, name: "d" }, server: "C", online: false, voice: null });
    const root = friendPresence([bob], [{ ...servers[0]!, voiceName: () => null }]);
    expect(root[0]!.voice).toBeNull();
  });
});
