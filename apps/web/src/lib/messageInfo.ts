import { formatDay, formatTime, type TimeFormat } from "./chat";

// A message's info (see MessageInfo): when it was sent, edited and pinned.

export const longDate = (at: number, locale?: string, clock?: TimeFormat) => `${formatDay(at, locale)} at ${formatTime(at, locale, clock)}`;

export type InfoRow = { label: string; value: string };
export const infoRows = (m: { at: number; edited?: number; pinned?: number }, locale?: string, clock?: TimeFormat): InfoRow[] => [
  { label: "Sent", value: longDate(m.at, locale, clock) },
  ...(m.edited ? [{ label: "Edited", value: longDate(m.edited, locale, clock) }] : []),
  ...(m.pinned ? [{ label: "Pinned", value: longDate(m.pinned, locale, clock) }] : []),
];
