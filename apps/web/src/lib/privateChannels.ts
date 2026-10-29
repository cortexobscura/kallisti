import type { Registered } from "@kallisti/protocol";
import type { ChannelKind } from "./channelForm";
import type { Speaker } from "./listenOnly";

// Private channels (built-in server): only their members, admins and
// moderators see them. The server keeps them on the Mumble channel: a
// voice channel itself, or a text channel's mirror. Members are
// registered users (also guest accounts, which are registered).

export type Member = Speaker;

type Mirrored = { id: number; mumble?: number };

// The Mumble channel that holds a channel's access (null: none, a text
// channel without a mirror): a voice channel itself, a text channel's
// mirror, a category's channel.
export const accessChannel = (kind: ChannelKind, id: number, textChannels: readonly Mirrored[], categories: readonly Mirrored[] = []) =>
  (kind === "voice" ? id : (kind === "text" ? textChannels : categories).find((c) => c.id === id)?.mumble) ?? null;

type Access = { private: Readonly<Record<number, readonly Member[]>>; inherited: Readonly<Record<number, readonly Member[]>> };
type Channels = { channels: readonly (Mirrored & { category?: number })[]; categories: readonly Mirrored[] };

// Who reads a channel (undefined: not private, or no list for us): a
// private one's members and those by its category; a channel in a
// private category, the category's.
export const readersOf = (text: Access & Channels, kind: "text" | "voice", id: number): Member[] | undefined => {
  const access = kind === "voice" ? id : text.channels.find((c) => c.id === id)?.mumble;
  if (access !== undefined && access in text.private) return [...text.private[access]!, ...(text.inherited[access] ?? [])];
  const cat = kind === "text" ? categoryAccess(text, text.channels.find((c) => c.id === id)?.category) : null;
  return cat !== null && cat in text.private ? [...text.private[cat]!] : undefined;
};

const categoryAccess = (text: Channels, category: number | undefined) =>
  category === undefined ? null : (text.categories.find((c) => c.id === category)?.mumble ?? null);

// A lock on a text channel: private, or in a private category.
export const isLocked = (text: { private: Readonly<Record<number, unknown>> } & Channels, c: Mirrored & { category?: number }) =>
  (c.mumble !== undefined && c.mumble in text.private) || (categoryAccess(text, c.category) ?? -1) in text.private;

const byName = (a: Member, b: Member) => a.name.localeCompare(b.name);
const ids = (ms: readonly Member[]) => ms.map((m) => m.id).sort((a, b) => a - b);

// Who may become a member: registered users who are not members yet, by
// name. SuperUser (id 0) sees all anyway.
export const memberCandidates = (registered: readonly Registered[] | null, members: readonly Member[]): Member[] =>
  (registered ?? [])
    .filter((r) => r.userId !== 0 && !members.some((m) => m.id === r.userId))
    .map((r) => ({ id: r.userId, name: r.name }))
    .sort(byName);

// What to ask the server for after the members dialog (null: no change).
export const membersAction = (channel: number, before: readonly Member[], after: readonly Member[]) =>
  ids(before).join() === ids(after).join() ? null : { setPrivate: channel, on: true, users: ids(after) };

// What to ask for after the private switch changed (null: no change).
export const privateAction = (channel: number, before: boolean, after: boolean, members: readonly Member[]) =>
  before === after ? null : { setPrivate: channel, on: after, users: after ? ids(members) : [] };
