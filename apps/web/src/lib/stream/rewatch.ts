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
