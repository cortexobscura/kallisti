import { addServer, loadServers, removeServer, saveServers, storedOwner } from "./serverList";

// What the hub needs of a server's session (see lib/serverSession).
type Closable = { leave: () => void; dispose: () => void; close: () => void };

export type HubDeps<S extends Closable> = {
  // localStorage: the list and the tokens live on the device.
  storage: () => Storage | undefined;
  // The home server's origin (this site, or the desktop app's choice).
  home: () => string;
  // Whose they are: the home login (server and account); null: none, or a
  // guest (who adds none: nothing of theirs survives on a shared device).
  owner: () => string | null;
  // A server's session token on this device.
  tokens: (origin: string) => { get: () => string | null; set: (token: string | null) => void };
  // A started session for a server, logged in as name (with its token).
  open: (origin: string, name: string) => S;
  // Logs out there with this token (best effort, after it is forgotten).
  revoke: (origin: string, token: string) => Promise<void>;
};

export type HubServer<S> = { origin: string; name: string; session: S };

// The servers added to the rail (several servers in one client): all
// connected, one shown (null: the home server). The list and tokens are
// on the device and end with the home login (endAll); later the home
// account keeps the list (meta/issues/home-server-list.md).
export const createServerHub = <S extends Closable>(deps: HubDeps<S>) => {
  let servers = $state.raw<HubServer<S>[]>([]);
  let shown = $state<string | null>(null);
  let restored = false;
  // Whose the open ones are (see follow).
  let live: string | null = null;

  const save = (owner: string, list: ReturnType<typeof loadServers>) => saveServers(deps.storage(), owner, list);
  // Forgets the token at once; logs out there with it afterwards, so a
  // slow answer cannot touch a later login's token.
  const forget = (origin: string) => {
    const store = deps.tokens(origin);
    const token = store.get();
    store.set(null);
    if (token) void deps.revoke(origin, token).catch(() => {});
  };
  const end = (s: HubServer<S>) => {
    s.session.leave();
    s.session.dispose();
  };
  // keepRestored: the current login stays (restore found another's list).
  const endAll = (keepRestored: boolean) => {
    const all = servers;
    servers = [];
    shown = null;
    restored = keepRestored;
    for (const { origin } of loadServers(deps.storage())) forget(origin);
    for (const s of all) {
      forget(s.origin);
      end(s);
    }
    saveServers(deps.storage(), "", []);
  };

  return {
    get servers(): readonly HubServer<S>[] {
      return servers;
    },
    get shown() {
      return shown;
    },
    show: (origin: string | null) => {
      shown = origin !== null && servers.some((s) => s.origin === origin) ? origin : null;
    },
    // A server logged in as name with this token (from the add dialog);
    // false: it is the home server or in the list (the first login stays).
    add: (origin: string, name: string, token: string) => {
      const owner = deps.owner();
      if (owner === null || origin === deps.home() || servers.some((s) => s.origin === origin)) return false;
      deps.tokens(origin).set(token);
      servers = [...servers, { origin, name, session: deps.open(origin, name) }];
      save(owner, addServer(loadServers(deps.storage(), owner), { origin, name }, deps.home()));
      shown = origin;
      return true;
    },
    // Out of the rail: logged out there, its token and entry gone.
    remove: (origin: string) => {
      const s = servers.find((x) => x.origin === origin);
      if (shown === origin) shown = null;
      const owner = deps.owner();
      if (owner !== null) save(owner, removeServer(loadServers(deps.storage(), owner), origin));
      forget(origin);
      if (!s) return;
      servers = servers.filter((x) => x !== s);
      end(s);
    },
    // The saved ones, once per home login (a token that no longer works
    // shows as an error there). Another login's end: logged out there,
    // tokens and list gone (a login that expired while the tab was closed
    // never ran endAll).
    restore: () => {
      const owner = deps.owner();
      if (restored || owner === null) return;
      restored = true;
      if (storedOwner(deps.storage()) !== owner) return endAll(true);
      const opened = new Set(servers.map((s) => s.origin));
      const fresh = loadServers(deps.storage(), owner).filter((s) => s.origin !== deps.home() && !opened.has(s.origin));
      servers = [...servers, ...fresh.map(({ origin, name }) => ({ origin, name, session: deps.open(origin, name) }))];
    },
    // The home login ended: all of them end, logged out there, list and
    // tokens gone from the device.
    endAll: () => endAll(false),
    // The home login now, on every change once it is known (also null: no
    // account, or a guest): another than the open ones' owner ends them at
    // once (a cookie that expired in an open tab runs no logout), and
    // another login's saved ones end before anything opens.
    follow: (owner: string | null) => {
      if (live !== null && live !== owner) endAll(false);
      live = owner;
      if (storedOwner(deps.storage()) !== owner && loadServers(deps.storage()).length > 0) endAll(restored);
    },
    // The app goes away: connections close; list and tokens stay.
    close: () => {
      for (const s of servers) s.session.close();
    },
  };
};
