import { addServer, loadServers, removeServer, saveServers } from "./serverList";

// What the hub needs of a server's session (see lib/serverSession).
type Closable = { leave: () => void; dispose: () => void; close: () => void };

export type HubDeps<S extends Closable> = {
  // localStorage: the list and the tokens live on the device.
  storage: () => Storage | undefined;
  // The home server's origin (this site, or the desktop app's choice).
  home: () => string;
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

  const save = (list: ReturnType<typeof loadServers>) => saveServers(deps.storage(), list);
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
      if (origin === deps.home() || servers.some((s) => s.origin === origin)) return false;
      deps.tokens(origin).set(token);
      servers = [...servers, { origin, name, session: deps.open(origin, name) }];
      save(addServer(loadServers(deps.storage()), { origin, name }, deps.home()));
      shown = origin;
      return true;
    },
    // Out of the rail: logged out there, its token and entry gone.
    remove: (origin: string) => {
      const s = servers.find((x) => x.origin === origin);
      if (shown === origin) shown = null;
      save(removeServer(loadServers(deps.storage()), origin));
      forget(origin);
      if (!s) return;
      servers = servers.filter((x) => x !== s);
      end(s);
    },
    // The saved ones, once per home login (a token that no longer works
    // shows as an error there, with "Remove").
    restore: () => {
      if (restored) return;
      restored = true;
      const opened = new Set(servers.map((s) => s.origin));
      const fresh = loadServers(deps.storage()).filter((s) => s.origin !== deps.home() && !opened.has(s.origin));
      servers = [...servers, ...fresh.map(({ origin, name }) => ({ origin, name, session: deps.open(origin, name) }))];
    },
    // The home login ended: all of them end, logged out there, list and
    // tokens gone from the device.
    endAll: () => {
      const all = servers;
      servers = [];
      shown = null;
      restored = false;
      for (const { origin } of loadServers(deps.storage())) forget(origin);
      for (const s of all) {
        forget(s.origin);
        end(s);
      }
      save([]);
    },
    // The app goes away: connections close; list and tokens stay.
    close: () => {
      for (const s of servers) s.session.close();
    },
  };
};
