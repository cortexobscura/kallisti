import type { PushPrefs } from "./push";

// Push from added servers (see lib/push syncServerPush): on while the home
// server's is on in this browser (turned on or off here, or in another
// tab), each with its own chats' settings; servers that went (removed,
// logged out) lose theirs, once the list is known (before it, all would
// look gone).
type Server = { origin: string; name: string; token: string | null };

export const createServerPush = (deps: {
  // The home account (null: none, a guest).
  name: () => string | null;
  // The home account's list came (see lib/serverHub loaded).
  loaded: () => boolean;
  servers: () => readonly Server[];
  isOn: (name: string) => Promise<boolean>;
  sync: (server: { origin: string; token: string; name: string }, homeOn: boolean, prefs: PushPrefs) => Promise<void>;
  drop: (keep: readonly string[]) => Promise<void>;
}) => {
  let homeOn = $state(false);
  const dispose = $effect.root(() => {
    $effect(() => {
      const name = deps.name();
      const check = () => {
        if (name) void deps.isOn(name).then((on) => (homeOn = on));
        else homeOn = false;
      };
      check();
      const other = (e: StorageEvent) => e.key === "pushOwner" && check();
      addEventListener("kallisti-push", check);
      addEventListener("storage", other);
      return () => {
        removeEventListener("kallisti-push", check);
        removeEventListener("storage", other);
      };
    });
    $effect(() => {
      const keep = deps.servers().flatMap((s) => (s.token ? [s.origin] : []));
      if (deps.loaded()) void deps.drop(keep);
    });
  });
  return {
    // For an added server's client (see Client pushSync).
    syncFor: (s: Server) => (prefs: PushPrefs) => {
      if (s.token) void deps.sync({ origin: s.origin, token: s.token, name: s.name }, homeOn, prefs);
    },
    dispose,
  };
};
