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

// What "Copy link" on a channel makes: a one-use invite link to its
// private channel or category (access: the Mumble channel that holds its
// members) for who may make invites, else a plain link (a private
// channel's works for its members only); null: none (a category, for
// others).
export const linkPlan = (
  site: string,
  kind: ChannelKind,
  id: number,
  { access, invites }: { access: number | null; invites: boolean },
): { invite: number } | { url: string } | null =>
  invites && access !== null
    ? { invite: access }
    : kind === "category"
      ? null
      : { url: kind === "voice" ? voiceLink(site, id) : channelLink(site, id) };
