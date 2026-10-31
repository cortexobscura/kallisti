import type { Online } from "./members";

// A friend is an account on one server (this one or another): its origin,
// its registered user id there, and the name last seen. The same person on
// two servers is two friends. Kept with the home account (the server's
// /api/friends), else on the device.
export type Friend = { origin: string; user: number; name: string };

const webOrigin = (s: string) => {
  try {
    const u = new URL(s);
    return (u.protocol === "https:" || u.protocol === "http:") && u.origin === s;
  } catch {
    return false;
  }
};
const isFriendEntry = (v: unknown): v is Friend => {
  const f = v as Record<string, unknown> | null;
  return !!f && typeof f.origin === "string" && webOrigin(f.origin) && Number.isInteger(f.user) && (f.user as number) >= 0 && typeof f.name === "string" && f.name !== "";
};
// Entries that are not friends, and repeats, left out.
export const parseFriends = (v: unknown): Friend[] =>
  Array.isArray(v)
    ? v.filter(isFriendEntry).reduce<Friend[]>((list, { origin, user, name }) => (isFriend(list, origin, user) ? list : [...list, { origin, user, name }]), [])
    : [];
// The list kept on the device (without a home account).
export const parseDeviceFriends = (raw: string | null): Friend[] => {
  try {
    return parseFriends(JSON.parse(raw ?? "[]"));
  } catch {
    return [];
  }
};

const same = (f: Friend, origin: string, user: number) => f.origin === origin && f.user === user;
export const isFriend = (list: readonly Friend[], origin: string, user: number) => list.some((f) => same(f, origin, user));
// A known one: the new name, the same place (as the server does).
export const withFriend = (list: readonly Friend[], f: Friend): Friend[] =>
  isFriend(list, f.origin, f.user) ? list.map((x) => (same(x, f.origin, f.user) ? f : x)) : [...list, f];
export const withoutFriend = (list: readonly Friend[], origin: string, user: number): Friend[] => list.filter((f) => !same(f, origin, user));

// What the client sees of one of its servers: who is online (user id to
// session), where sessions are, and a voice channel's name (null: none to
// show, such as the root).
export type FriendServer = {
  origin: string;
  name: string;
  online: Online;
  users: Readonly<Record<number, { channelId: number }>>;
  voiceName: (id: number) => string | null;
};
export type FriendRow = { friend: Friend; server: string | null; online: boolean; voice: string | null };

// Where friends are, from the servers the user is connected to (nothing
// else: no server tells another). Online ones first, each group in the
// list's order.
export const friendPresence = (list: readonly Friend[], servers: readonly FriendServer[]): FriendRow[] => {
  const rows = list.map((friend): FriendRow => {
    const s = servers.find((x) => x.origin === friend.origin);
    const session = s?.online[friend.user];
    const channel = session === undefined ? undefined : s!.users[session]?.channelId;
    return { friend, server: s?.name ?? null, online: session !== undefined, voice: channel === undefined ? null : s!.voiceName(channel) };
  });
  return [...rows.filter((r) => r.online), ...rows.filter((r) => !r.online)];
};
