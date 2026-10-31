import { withFriend, withoutFriend, type Friend } from "./friends";

type Remote = { load: () => Promise<Friend[]>; add: (f: Friend) => Promise<void>; remove: (origin: string, user: number) => Promise<void> };

// The friends list (see lib/friends) of its owner: a home account (kept on
// its home server, remote) or, null, this device. Another owner empties it
// at once (no account sees another's), and a change made for an earlier
// owner (a dialog left open over a logout) is not taken. Changes show at
// once and end loads still running (their answer is older); writes go one
// after the other; a refused one is told and brings the server's list back.
export const createFriends = (deps: {
  owner: () => string | null;
  remote: (owner: string) => Remote;
  device: { get: () => Friend[]; set: (list: Friend[]) => void };
  onError?: () => void;
}) => {
  let state = $state.raw<{ of: string | null | undefined; list: Friend[] }>({ of: undefined, list: [] });
  let loads = 0;
  let writes = Promise.resolve();
  const sync = async () => {
    const owner = deps.owner();
    if (owner !== state.of) state = { of: owner, list: [] };
    const mine = ++loads;
    if (owner === null) return void (state = { of: null, list: deps.device.get() });
    try {
      const list = await deps.remote(owner).load();
      if (mine === loads && state.of === owner) state = { of: owner, list };
    } catch {
      // Kept as it is: the next sync tries again.
    }
  };
  const change = (next: (list: Friend[]) => Friend[], send: (r: Remote) => Promise<void>) => {
    const owner = deps.owner();
    if (owner !== state.of) return;
    loads++;
    state = { of: owner, list: next(state.list) };
    if (owner === null) return deps.device.set(state.list);
    const remote = deps.remote(owner);
    writes = writes
      .then(() => send(remote))
      .catch(() => {
        deps.onError?.();
        void sync();
      });
  };
  return {
    get list() {
      return state.list;
    },
    sync,
    add: (f: Friend) => change((list) => withFriend(list, f), (r) => r.add(f)),
    remove: (origin: string, user: number) => change((list) => withoutFriend(list, origin, user), (r) => r.remove(origin, user)),
  };
};
