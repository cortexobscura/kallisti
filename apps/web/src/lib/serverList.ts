import { safeGet, safeSet } from "./storage";

// The Kallisti servers added on this device (several servers in one
// client), by origin with the name logged in there, in the rail's order,
// for one owner: the home login that added them (another login on this
// device does not get them, see lib/serverHub). The home server (this site, or
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

// id: the login that keeps the list (one per login, shared by its tabs;
// see lib/serverHub); at: when it was saved.
type Stored = { owner: string; id: string; at: number; servers: SavedServer[] };
const read = (storage: Storage | undefined): Stored => {
  try {
    const v = JSON.parse(safeGet(storage, KEY) ?? "null") as { owner?: unknown; id?: unknown; at?: unknown; servers?: unknown } | unknown[] | null;
    // Before owners: an array (no owner: ended, never restored).
    if (Array.isArray(v)) return { owner: "", id: "", at: 0, servers: v.map((x) => (typeof x === "string" ? { origin: x, name: "?" } : x)).filter(isSaved) };
    const text = (x: unknown) => (typeof x === "string" ? x : "");
    const servers = Array.isArray(v?.servers) ? v.servers.filter(isSaved).map(({ origin, name }) => ({ origin, name })) : [];
    return { owner: text(v?.owner), id: text(v?.id), at: typeof v?.at === "number" ? v.at : 0, servers };
  } catch {
    return { owner: "", id: "", at: 0, servers: [] };
  }
};

// The saved ones; with an owner, only if they are that login's.
export const loadServers = (storage: Storage | undefined, owner?: string): SavedServer[] => {
  const saved = read(storage);
  return owner === undefined || saved.owner === owner ? saved.servers : [];
};

export const storedOwner = (storage: Storage | undefined) => read(storage).owner;
export const storedId = (storage: Storage | undefined) => read(storage).id;
export const storedAt = (storage: Storage | undefined) => read(storage).at;

export const saveServers = (storage: Storage | undefined, owner: string, list: readonly SavedServer[], at = Date.now(), id = "") =>
  void safeSet(storage, KEY, JSON.stringify({ owner, id, at, servers: list }));

const local = (host: string) => host === "localhost" || host === "127.0.0.1" || host === "[::1]";

// A host without a trailing dot (one spelling of a name).
const host = (url: URL) => url.hostname.replace(/\.$/, "");

// The origin of another server from what the user typed: https (http only
// on this computer, for development), its port kept, any path dropped.
// Not the home server, and (homeCookie: the web app, whose home login is a
// cookie) not another port of its host: browsers send the home cookie to
// every port of a host.
export const serverOrigin = (input: string, home: string, homeCookie = true): { origin: string } | { error: string } => {
  const text = input.trim();
  const invalid = { error: "Enter a server address, for example chat.example.org" };
  let url: URL;
  try {
    url = new URL(/^[a-z][a-z0-9+.-]*:\/\//i.test(text) ? text : `https://${text}`);
  } catch {
    return invalid;
  }
  if (!text || !host(url) || (url.protocol !== "https:" && !(url.protocol === "http:" && local(host(url))))) return invalid;
  const origin = `${url.protocol}//${host(url)}${url.port ? `:${url.port}` : ""}`;
  let homeUrl: URL | null = null;
  try {
    homeUrl = new URL(home);
  } catch {
    // No home origin to compare.
  }
  const homeOrigin = homeUrl && `${homeUrl.protocol}//${host(homeUrl)}${homeUrl.port ? `:${homeUrl.port}` : ""}`;
  if (origin === homeOrigin) return { error: "This is your home server." };
  if (homeCookie && homeUrl && host(homeUrl) === host(url)) return { error: "Another port of your home server's host cannot be added: it would see your login there." };
  return { origin };
};
