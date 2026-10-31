// The member list's collapsed groups, by a key of their own ("friends",
// "here", "online", "offline", a reader group's title), on this device for
// all servers.
export const toggleGroup = (collapsed: readonly string[], key: string): string[] =>
  collapsed.includes(key) ? collapsed.filter((k) => k !== key) : [...collapsed, key];

export const parseCollapsedGroups = (raw: string | null): string[] => {
  try {
    const data: unknown = JSON.parse(raw ?? "[]");
    return Array.isArray(data) ? data.filter((k): k is string => typeof k === "string") : [];
  } catch {
    return [];
  }
};
