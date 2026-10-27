import { safeGet, safeSet } from "./storage";

// The Kallisti servers added on this device (several servers in one
// client), by origin, in the rail's order. The home server (this site, or
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

export const addServer = (list: readonly string[], origin: string, home: string) =>
  origin === home || list.includes(origin) ? [...list] : [...list, origin];

export const removeServer = (list: readonly string[], origin: string) => list.filter((o) => o !== origin);

export const loadServers = (storage: Storage | undefined): string[] => {
  try {
    const v: unknown = JSON.parse(safeGet(storage, KEY) ?? "[]");
    return Array.isArray(v) ? v.filter(isOrigin) : [];
  } catch {
    return [];
  }
};

export const saveServers = (storage: Storage | undefined, list: readonly string[]) => void safeSet(storage, KEY, JSON.stringify(list));
