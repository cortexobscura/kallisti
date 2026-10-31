import type { ChannelNode } from "@kallisti/protocol";

// A user's own channel list on a server: hidden channels and an own order,
// of text channels (their ids) and voice channels (their Mumble ids). Kept
// with the account on that server (the server's "channel-list"), else on
// the device.
export type ListKind = "text" | "voice";
type Ids = Readonly<Record<ListKind, readonly number[]>>;
export type ChannelList = { readonly hidden: Ids; readonly order: Ids };

export const noChannelList: ChannelList = { hidden: { text: [], voice: [] }, order: { text: [], voice: [] } };

const idsOf = (v: unknown): number[] => (Array.isArray(v) ? v.filter((id) => Number.isInteger(id)) : []);
const kinds = (v: unknown): Ids => {
  const o = v && typeof v === "object" ? (v as Record<string, unknown>) : {};
  return { text: idsOf(o.text), voice: idsOf(o.voice) };
};
export const parseChannelList = (v: unknown): ChannelList => {
  const o = v && typeof v === "object" ? (v as Record<string, unknown>) : {};
  return { hidden: kinds(o.hidden), order: kinds(o.order) };
};

// Those in the own order first, in its order; the others after, in the
// server's order (new channels go to the end of their group).
export const inOwnOrder = <T>(items: readonly T[], idOf: (t: T) => number, order: readonly number[]): T[] => {
  const place = new Map(order.map((id, i) => [id, i]));
  const at = (t: T) => place.get(idOf(t)) ?? order.length;
  return items.map((t, i) => ({ t, i })).sort((a, b) => at(a.t) - at(b.t) || a.i - b.i).map(({ t }) => t);
};

// A group's text channels as the user sees them. kept: shown even when
// hidden (the open channel, one with a mention of the user).
export const ownChannels = <T extends { id: number }>(
  channels: readonly T[],
  hidden: ReadonlySet<number>,
  order: readonly number[],
  kept: (id: number) => boolean,
): T[] => inOwnOrder(channels, (c) => c.id, order).filter((c) => !hidden.has(c.id) || kept(c.id));

// The voice channels as the user sees them: a hidden one goes with its
// subchannels (unless kept: the user is in it, or under it); each level
// in the own order.
export const ownTree = (node: ChannelNode, list: ChannelList, kept: (id: number) => boolean): ChannelNode => {
  const hidden = new Set(list.hidden.voice);
  const keeps = (n: ChannelNode): boolean => kept(n.channel.id) || n.children.some(keeps);
  const walk = (n: ChannelNode): ChannelNode => ({
    channel: n.channel,
    children: inOwnOrder(n.children, (c) => c.channel.id, list.order.voice)
      .filter((c) => !hidden.has(c.channel.id) || keeps(c))
      .map(walk),
  });
  return walk(node);
};

export const toggleHidden = (list: ChannelList, kind: ListKind, id: number): ChannelList => {
  const now = list.hidden[kind];
  return { ...list, hidden: { ...list.hidden, [kind]: now.includes(id) ? now.filter((x) => x !== id) : [...now, id] } };
};

// Moves a channel one place up (-1) or down (1) among its group (ids: the
// group as shown); the group's new order goes to the end of the own order.
export const moveOwn = (list: ChannelList, kind: ListKind, ids: readonly number[], id: number, by: -1 | 1): ChannelList => {
  const i = ids.indexOf(id);
  const j = i + by;
  if (i < 0 || j < 0 || j >= ids.length) return list;
  const next = [...ids];
  [next[i], next[j]] = [next[j]!, next[i]!];
  const group = new Set(ids);
  return { ...list, order: { ...list.order, [kind]: [...list.order[kind].filter((x) => !group.has(x)), ...next] } };
};

// Sets the account's list on the server (see the server's set-channel-list).
export const channelListRequest = (list: ChannelList) => JSON.stringify({ type: "set-channel-list", list });

// Without an account: one list per server (its site), on the device.
export type DeviceLists = Readonly<Record<string, ChannelList>>;
export const parseDeviceLists = (raw: string | null): DeviceLists => {
  let data: unknown;
  try {
    data = JSON.parse(raw ?? "null");
  } catch {
    return {};
  }
  if (!data || typeof data !== "object" || Array.isArray(data)) return {};
  return Object.fromEntries(Object.entries(data).map(([server, v]) => [server, parseChannelList(v)]));
};
export const deviceListOf = (all: DeviceLists, server: string): ChannelList => all[server] ?? noChannelList;
export const setDeviceList = (all: DeviceLists, server: string, list: ChannelList): DeviceLists => ({ ...all, [server]: list });

// The ids shown beside a text channel (its group as shown), for moves.
export const shownBeside = (groups: readonly { channels: readonly { id: number }[] }[], id: number): number[] =>
  groups.find((g) => g.channels.some((c) => c.id === id))?.channels.map((c) => c.id) ?? [];

// A voice channel and its siblings as shown (none for the root).
export const voiceBeside = (tree: ChannelNode, id: number): number[] => {
  const ids = tree.children.map((c) => c.channel.id);
  if (ids.includes(id)) return ids;
  for (const c of tree.children) {
    const found = voiceBeside(c, id);
    if (found.length) return found;
  }
  return [];
};

// Shown even when hidden: the open channel, one with a mention of the user.
export const keepShown =
  (current: number | null, mentions: Readonly<Record<number, number>>) =>
  (id: number): boolean =>
    id === current || (mentions[id] ?? 0) > 0;

// The sidebar as the user sees it: the text groups and the voice tree.
export const ownView = <G extends { channels: readonly { id: number }[] }>(
  groups: readonly G[],
  tree: ChannelNode | null,
  list: ChannelList,
  keep: { text: (id: number) => boolean; voice: (id: number) => boolean },
): { groups: G[]; tree: ChannelNode | null } => {
  const hidden = new Set(list.hidden.text);
  return {
    groups: groups.map((g) => ({ ...g, channels: ownChannels(g.channels, hidden, list.order.text, keep.text) })),
    tree: tree && ownTree(tree, list, keep.voice),
  };
};

// Without channels that are gone (a new channel can get a deleted one's id;
// and the lists must stay under the server's limit). Only with the
// channels known (after ready), or it would empty the list.
export const pruned = (list: ChannelList, text: ReadonlySet<number>, voice: ReadonlySet<number>): ChannelList => {
  const keep = (ids: Ids): Ids => ({ text: ids.text.filter((id) => text.has(id)), voice: ids.voice.filter((id) => voice.has(id)) });
  return { hidden: keep(list.hidden), order: keep(list.order) };
};
