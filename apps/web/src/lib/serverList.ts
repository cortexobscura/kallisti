import { safeGet, safeSet } from "./storage";

// The Kallisti servers added on this device (several servers in one
// client), by origin with the name logged in there, in the rail's order. The home server (this site, or
// the desktop app's chosen one) is not in it. Later the home account keeps
// the list (meta/issues/home-server-list.md).

const KEY = "servers";

// An http(s) origin, and only that (no path, no other scheme).
const isOrigin = (v: unknown): v is string => {
  if (typeof v !== "string") return false;
  try {
    const u = new URL(v);
    return (u.protocol === "https:" || u.protocol === "http:") && u.origin === v;
  } catch {
    return false;
  }
};

export type SavedServer = { origin: string; name: string };

const isSaved = (v: unknown): v is SavedServer =>
  !!v && typeof v === "object" && isOrigin((v as SavedServer).origin) && typeof (v as SavedServer).name === "string" && (v as SavedServer).name !== "";

export const addServer = (list: readonly SavedServer[], server: SavedServer, home: string) =>
  server.origin === home || list.some((s) => s.origin === server.origin) ? [...list] : [...list, server];

export const removeServer = (list: readonly SavedServer[], origin: string) => list.filter((s) => s.origin !== origin);

export const loadServers = (storage: Storage | undefined): SavedServer[] => {
  try {
    const v: unknown = JSON.parse(safeGet(storage, KEY) ?? "[]");
    return Array.isArray(v) ? v.filter(isSaved).map(({ origin, name }) => ({ origin, name })) : [];
  } catch {
    return [];
  }
};

export const saveServers = (storage: Storage | undefined, list: readonly SavedServer[]) => void safeSet(storage, KEY, JSON.stringify(list));

const local = (host: string) => host === "localhost" || host === "127.0.0.1" || host === "[::1]";

// The origin of another server from what the user typed: https (http only
// on this computer, for development), its port kept, any path dropped.
// Not the home server, and not another port of its host: browsers send the
// home cookie to every port of a host.
export const serverOrigin = (input: string, home: string): { origin: string } | { error: string } => {
  const text = input.trim();
  let url: URL;
  try {
    url = new URL(/^[a-z][a-z0-9+.-]*:/i.test(text) ? text : `https://${text}`);
  } catch {
    return { error: "Enter a server address, for example chat.example.org" };
  }
  if (!text || !url.hostname || (url.protocol !== "https:" && !(url.protocol === "http:" && local(url.hostname)))) {
    return { error: "Enter a server address, for example chat.example.org" };
  }
  if (url.origin === home) return { error: "This is your home server." };
  try {
    if (new URL(home).hostname === url.hostname) return { error: "Another port of your home server's host cannot be added: it would see your login there." };
  } catch {
    // No home origin to compare.
  }
  return { origin: url.origin };
};
