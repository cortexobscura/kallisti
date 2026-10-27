// The tab title: the number of mentions and direct messages not read
// first, as in Discord ("(3) Kallisti").
export const tabTitle = (name: string, count: number) => (count > 0 ? `(${count > 99 ? "99+" : count}) ${name}` : name);

const sum = (r: Record<string | number, number>) => Object.values(r).reduce((a, n) => a + n, 0);

// A server's mark in the rail (several servers in one client), as in
// Discord: the number of mentions and direct messages not read, and a dot
// for other unread text.
export type RailBadge = { mentions: number; unread: boolean };
export const railBadge = (mentions: Record<number, number>, directs: Record<string, number>, unreadText: ReadonlySet<number>): RailBadge => ({
  mentions: sum(mentions) + sum(directs),
  unread: unreadText.size > 0,
});
