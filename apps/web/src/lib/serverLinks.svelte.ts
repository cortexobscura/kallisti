import { inviteOf, inviteRoute } from "./invites";
import { parseMessageLink } from "./messageLink";

// Links from elsewhere (several servers in one client): an invite clicked
// in any client or opened with ?join=, a push notice's message (?open=).
// That server goes on screen (added first, or its login first, if needed)
// and its client takes the link (see Client link): an invite is redeemed,
// a message opens. The links end with their home login.
export const createServerLinks = (deps: {
  servers: () => readonly { origin: string; name: string; session: unknown }[];
  home: () => string;
  owner: () => string | null;
  select: (id: string) => void;
  // The add dialog: for a server of the list without a login here, or a
  // new address.
  add: (target: { server: { origin: string; name: string } } | { address: string }) => void;
}) => {
  let pending = $state<Record<string, string>>({});
  let of = deps.owner();
  const dispose = $effect.root(() => {
    $effect(() => {
      const owner = deps.owner();
      if (owner === of) return;
      of = owner;
      pending = {};
    });
  });
  const homeOrigin = () => new URL(deps.home()).origin;
  return {
    // An invite link; false: not one, or a guest (who adds no servers).
    openInvite: (url: string): boolean => {
      const inv = inviteOf(url);
      if (!inv) return false;
      const known = deps.servers().find((s) => s.origin === inv.origin);
      const route = inviteRoute(inv.origin, homeOrigin(), known && { session: !!known.session }, deps.owner() === null);
      if (route === null) return false;
      const id = route === "home" ? "" : inv.origin;
      pending[id] = url;
      if (route !== "add") deps.select(id);
      if (route === "login") deps.add({ server: { origin: known!.origin, name: known!.name } });
      if (route === "add") deps.add({ address: inv.origin });
      return true;
    },
    // A message link of a server here (with a session); others do nothing.
    openMessageLink: (url: string) => {
      let origin: string;
      try {
        origin = new URL(url).origin;
      } catch {
        return;
      }
      const id = origin === homeOrigin() ? "" : deps.servers().find((s) => s.origin === origin && s.session)?.origin;
      if (id === undefined || !parseMessageLink(url, origin)) return;
      pending[id] = url;
      deps.select(id);
    },
    // The link for a server's client ("": home), until it is used.
    linkFor: (id: string): string | null => pending[id] ?? null,
    used: (id: string) => void delete pending[id],
    dispose,
  };
};
