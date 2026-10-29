import type { ChannelKind } from "./channelForm";

// Links to a stored message ("Copy link"): <server>/channels/<chat>/<id>,
// to a chat: <server>/channels/<chat>, and to a voice channel:
// <server>/voice/<id> (joined after a question). The server answers any
// such path with the app, which opens the chat and jumps to the message.
export const messageLink = (origin: string, chat: number, message: number) => `${origin}/channels/${chat}/${message}`;
export const channelLink = (origin: string, chat: number) => `${origin}/channels/${chat}`;
export const voiceLink = (origin: string, channel: number) => `${origin}/voice/${channel}`;

export type ChannelLink = { chat: number; message?: number } | { voice: number };

// null: not a message, chat or voice channel link to this server (origin).
export const parseMessageLink = (url: string, origin: string): ChannelLink | null => {
  let u: URL;
  try {
    u = new URL(url, origin);
  } catch {
    return null;
  }
  if (u.origin !== new URL(origin).origin) return null;
  const v = /^\/voice\/(\d+)$/.exec(u.pathname);
  if (v) return { voice: Number(v[1]) };
  const m = /^\/channels\/(\d+)(?:\/(\d+))?$/.exec(u.pathname);
  if (!m) return null;
  return m[2] ? { chat: Number(m[1]), message: Number(m[2]) } : { chat: Number(m[1]) };
};

// "Copy link": a plain link to a text or voice channel (a private one's
// works for who sees it); null: none (a category).
export const plainLink = (site: string, kind: ChannelKind, id: number): string | null =>
  kind === "category" ? null : kind === "voice" ? voiceLink(site, id) : channelLink(site, id);

// "Copy invite link": the private channel or category to invite to
// (access: the Mumble channel that holds its members) for who may make
// invites; null: none.
export const inviteTarget = ({ access, invites }: { access: number | null; invites: boolean }): number | null =>
  invites && access !== null ? access : null;
