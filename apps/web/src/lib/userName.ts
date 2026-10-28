// An account from another server is named as the fediverse writes it,
// name@home (see the server's names.Remote): its own part, and the host
// (shown faded, as Discord shows its numbers). host null: a name of here.
export const splitName = (name: string): { own: string; host: string | null } => {
  const i = name.lastIndexOf("@");
  const host = i > 0 ? name.slice(i + 1) : "";
  return /^[a-z0-9-]+(\.[a-z0-9-]+)+(:\d{1,5})?$/.test(host) ? { own: name.slice(0, i), host } : { own: name, host: null };
};
