import { safeGet, safeSet } from "../storage";
import { isCamera } from "./protocol";

// A watched stream that ended may come back from the same user under a new
// id: a stream from OBS goes on with its owner's next voice session (after a
// browser refresh, see the server's whip package). Within a while it is
// watched again by itself.

export const REWATCH_MS = 30_000;

export type Lost = { name: string; camera: boolean; until: number };

// The streams here (in our voice channel) to watch again, and what is still
// waited for.
export const rewatch = (lost: readonly Lost[], here: readonly { id: number; name: string; camera: boolean }[], now: number) => {
  const key = (name: string, camera: boolean) => `${name.toLowerCase()} ${camera}`;
  const back = new Map(here.map((s) => [key(s.name, s.camera), s.id]));
  const watch: number[] = [];
  const waiting: Lost[] = [];
  for (const l of lost) {
    if (now >= l.until) continue;
    const id = back.get(key(l.name, l.camera));
    if (id === undefined) waiting.push(l);
    else watch.push(id);
  }
  return { watch, lost: waiting };
};

// What the viewer watched, by stream id: the owner's first known name, the
// ones no longer watched (lost if the stream ends), and those waited for.
export type Watched = { names: ReadonlyMap<number, string>; dropped: ReadonlyMap<number, Lost>; lost: readonly Lost[] };
export const NOTHING_WATCHED: Watched = { names: new Map(), dropped: new Map(), lost: [] };

export type Seen = {
  here: readonly { id: number; name: string; camera: boolean; known: boolean }[]; // streams in our voice channel (known: the owner is in the user list)
  live: ReadonlySet<number>; // all streams
  watching: readonly number[];
  closed: ReadonlySet<number>; // closed by the viewer: never watched again by itself
  inVoice: boolean;
  now: number;
};

// One step: streams not here are not watched; a watched one that ended
// (left the stream list) is watched again when the same user's comes back.
export const follow = (w: Watched, s: Seen) => {
  const isHere = (id: number) => s.here.some((h) => h.id === id);
  const unwatch = s.watching.filter((id) => !isHere(id));
  const names = new Map(w.names);
  // The user list can lose the streamer before the stream list: the first
  // known name counts.
  for (const h of s.here) if (s.watching.includes(h.id) && h.known && !names.has(h.id)) names.set(h.id, h.name);
  const dropped = new Map(w.dropped);
  for (const [id, name] of names) {
    if (s.watching.includes(id) && isHere(id)) continue;
    names.delete(id);
    if (!s.closed.has(id)) dropped.set(id, { name, camera: isCamera(id), until: s.now + REWATCH_MS });
  }
  const lost = [...w.lost];
  for (const [id, l] of dropped) {
    if (!s.live.has(id) && s.inVoice) lost.push(l);
    if (!s.live.has(id) || s.now >= l.until) dropped.delete(id);
  }
  const back = rewatch(lost, s.here, s.now);
  return { watched: { names, dropped, lost: back.lost }, unwatch, watch: back.watch };
};

// The streams watched and those waited for, kept above the Client (new
// after each reconnect) and over a reload (see leaving, saveLost).
export type StreamMemory = { watching: { name: string; camera: boolean }[]; lost: Lost[] };

// The connection or the page goes (a reconnect makes a new Client, a reload
// a new page): the streams watched are lost from now, with those still
// waited for, one entry each.
export const leaving = (watching: readonly { name: string; camera: boolean }[], waited: readonly Lost[], now: number): Lost[] => {
  const key = (l: { name: string; camera: boolean }) => `${l.name.toLowerCase()} ${l.camera}`;
  const fresh = watching.map((w) => ({ name: w.name, camera: w.camera, until: now + REWATCH_MS }));
  const mine = new Set(fresh.map(key));
  return [...waited.filter((l) => l.until > now && !mine.has(key(l))), ...fresh];
};

// Over a reload: per tab (sessionStorage), as the place (see lib/place).
type Storage = Parameters<typeof safeGet>[0];
const isLost = (v: unknown): v is Lost =>
  typeof v === "object" && v !== null && typeof (v as Lost).name === "string" && typeof (v as Lost).camera === "boolean" && Number.isFinite((v as Lost).until);
export const saveLost = (storage: Storage, key: string, lost: readonly Lost[]) => safeSet(storage, key, JSON.stringify(lost));
export const loadLost = (storage: Storage, key: string, now: number): Lost[] => {
  try {
    const v: unknown = JSON.parse(safeGet(storage, key) ?? "[]");
    return Array.isArray(v) ? v.filter(isLost).filter((l) => l.until > now) : [];
  } catch {
    return [];
  }
};
