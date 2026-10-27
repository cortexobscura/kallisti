import { type Login, loadTokens, type SavedServer, saveTokens } from "./serverList";

// What the hub needs of a server's session (see lib/serverSession).
type Closable = { leave: () => void; dispose: () => void; close: () => void };

// The home account's list (see lib/accounts servers).
export type ListApi = {
  load: () => Promise<SavedServer[]>;
  add: (server: SavedServer) => Promise<void>;
  remove: (origin: string) => Promise<void>;
};

export type HubDeps<S extends Closable> = {
  // localStorage: the tokens live on the device.
  storage: () => Storage | undefined;
  // A server this client may connect to (not the home server; in the web
  // app not another port of its host: see serverOrigin).
  allowed: (origin: string) => boolean;
  // The home account's list; null: none (no login, a guest, a Mumble home).
  list: () => ListApi | null;
  // Whose the tokens here are: the home login (server and account); null:
  // none, or a guest.
  owner: () => string | null;
  // A started session for a server, logged in as name with this token.
  open: (origin: string, name: string, token: string) => S;
  // Logs out there with this token (best effort, after it is forgotten).
  revoke: (origin: string, token: string) => Promise<void>;
};

// session null: no token on this device (it asks for a login there).
export type HubServer<S> = { origin: string; name: string; session: S | null; token: string | null };

// The servers added to the rail (several servers in one client): the home
// account keeps the list, this device the tokens; all with a token are
// connected, one is shown (null: the home server). A home logout ends them
// here (logged out there, tokens gone); the list stays at home.
export const createServerHub = <S extends Closable>(deps: HubDeps<S>) => {
  let servers = $state.raw<HubServer<S>[]>([]);
  let shown = $state<string | null>(null);
  // The last list change the home server did not take.
  let error = $state<string | null>(null);
  // The home login the open ones and the tokens belong to.
  let live: string | null = null;
  let first = true;
  // Adds on their way home: a list loaded meanwhile does not have them yet.
  const pending = new Set<string>();
  // Changes here: a list loaded before one is older than this tab.
  let changes = 0;

  const tokens = () => {
    const t = loadTokens(deps.storage());
    return t.owner === live ? t.tokens : {};
  };
  // Another tab's newer login keeps the tokens here: a stale tab adds none.
  const mine = () => [live ?? "", ""].includes(loadTokens(deps.storage()).owner);
  const setToken = (origin: string, login: Login | null) => {
    const { [origin]: _, ...rest } = tokens();
    saveTokens(deps.storage(), live ?? "", login ? { ...rest, [origin]: login } : rest);
  };
  // With a token here: logged in as the name that token has.
  const entry = (origin: string, name: string, login = tokens()[origin]): HubServer<S> => ({
    origin,
    name: login?.name ?? name,
    token: login?.token ?? null,
    session: login ? deps.open(origin, login.name, login.token) : null,
  });
  // Logged out there, its token forgotten (if it is still that one).
  const forget = (origin: string, token: string) => {
    if (tokens()[origin]?.token === token) setToken(origin, null);
    void deps.revoke(origin, token).catch(() => {});
  };
  // Closed, logged out there, its token forgotten.
  const end = (s: HubServer<S>) => {
    s.session?.leave();
    s.session?.dispose();
    if (s.token) forget(s.origin, s.token);
  };
  const drop = (s: HubServer<S>) => {
    changes++;
    servers = servers.filter((x) => x !== s);
    if (shown === s.origin) shown = null;
    end(s);
  };
  // All tokens here end (logged out there) if they are of a login other
  // than keep.
  const endTokens = (keep: string | null) => {
    const stored = loadTokens(deps.storage());
    if (stored.owner === (keep ?? "")) return;
    for (const [origin, { token }] of Object.entries(stored.tokens)) void deps.revoke(origin, token).catch(() => {});
    saveTokens(deps.storage(), keep ?? "", {});
  };
  const closeAll = (each: (s: HubServer<S>) => void) => {
    const all = servers;
    servers = [];
    shown = null;
    for (const s of all) each(s);
  };
  // This login ends here: the open ones and its tokens (logged out there);
  // not tokens a newer login in another tab has now. Then owner's.
  const endLogin = (owner: string | null) => {
    const wasMine = mine();
    closeAll(end);
    if (wasMine) endTokens(owner);
    live = owner;
  };
  // undo: the new entry, dropped if home does not take it (not for a rename).
  const save = (server: SavedServer, undo?: HubServer<S>) => {
    const api = deps.list();
    if (!api) return;
    pending.add(server.origin);
    api.add(server).then(
      () => void pending.delete(server.origin),
      (err: unknown) => {
        pending.delete(server.origin);
        if (undo && servers.includes(undo)) drop(undo);
        error = `${new URL(server.origin).host} was not saved: ${err instanceof Error ? err.message : String(err)}`;
      },
    );
  };

  return {
    get servers(): readonly HubServer<S>[] {
      return servers;
    },
    get shown() {
      return shown;
    },
    get error() {
      return error;
    },
    show: (origin: string | null) => {
      shown = origin !== null && servers.some((s) => s.origin === origin) ? origin : null;
    },
    // The home account's list now: new ones open, ones removed elsewhere end
    // (also their tokens here from before), ones logged in by another tab
    // connect.
    sync: async () => {
      const api = deps.list();
      const owner = live;
      const at = changes;
      if (!api || owner === null) return;
      const list = await api.load().catch(() => null);
      if (list === null || live !== owner || changes !== at) return;
      const want = list.filter((x) => deps.allowed(x.origin));
      const kept = (origin: string) => pending.has(origin) || want.some((x) => x.origin === origin);
      for (const s of servers) if (!kept(s.origin)) drop(s);
      for (const [origin, { token }] of Object.entries(tokens())) if (!kept(origin)) forget(origin, token);
      servers = servers.map((s) => (s.session || !tokens()[s.origin] ? s : entry(s.origin, s.name)));
      const added = want.filter((x) => !servers.some((s) => s.origin === x.origin)).map((x) => entry(x.origin, x.name));
      servers = [...servers, ...added];
    },
    // A server logged in as name with this token (from the add dialog);
    // false: no account at home, the home server, or already in the rail.
    add: (origin: string, name: string, token: string) => {
      if (live === null || deps.owner() !== live || !mine() || !deps.allowed(origin) || servers.some((s) => s.origin === origin)) return false;
      changes++;
      const s = entry(origin, name, { token, name });
      setToken(origin, { token, name });
      servers = [...servers, s];
      shown = origin;
      error = null;
      save({ origin, name }, s);
      return true;
    },
    // A server of the list without a token here, logged in now.
    login: (origin: string, name: string, token: string) => {
      const old = servers.find((s) => s.origin === origin);
      if (!old || old.session || deps.owner() !== live || !mine()) return false;
      setToken(origin, { token, name });
      const s = entry(origin, name, { token, name });
      servers = servers.map((x) => (x === old ? s : x));
      if (name !== old.name) save({ origin, name });
      return true;
    },
    // Out of the rail and the list: logged out there, its token gone.
    remove: (origin: string) => {
      const s = servers.find((x) => x.origin === origin);
      if (!s) return;
      deps
        .list()
        ?.remove(origin)
        .catch((err: unknown) => void (error = `${new URL(origin).host} is still in your list: ${err instanceof Error ? err.message : String(err)}`));
      drop(s);
    },
    // The home login ended: all end here; the list stays at home.
    endAll: () => endLogin(null),
    // Another login is in this browser now (a stale tab): the open ones
    // close; the tokens are that login's.
    detach: () => {
      closeAll((s) => (s.session?.leave(), s.session?.dispose()));
      live = null;
    },
    // The home login now (also null: none, or a guest), from the first
    // known one on: another than the open ones' ends them. At the first,
    // tokens of another login end (one that ended while no tab was open).
    follow: (owner: string | null) => {
      if (first) {
        first = false;
        endTokens(owner);
        live = owner;
      } else if (owner !== live) endLogin(owner);
    },
    // The app goes away: connections close; tokens stay.
    close: () => {
      for (const s of servers) s.session?.close();
    },
  };
};
