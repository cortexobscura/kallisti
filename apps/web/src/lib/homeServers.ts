// Which other servers' accounts may log in here (a Kallisti server or a
// fediverse instance as their home; see the server's package homes): all,
// all but a block list, or only an allow list. Admins set it.

export type HomeServers = { mode: "all" | "block" | "allow"; hosts: string[] };

const MODES = ["all", "block", "allow"] as const;

export const parseHomeServers = (v: unknown): HomeServers | null => {
  const h = v as Partial<HomeServers> | null;
  if (!h || !MODES.includes(h.mode as HomeServers["mode"]) || !Array.isArray(h.hosts) || !h.hosts.every((x) => typeof x === "string")) return null;
  return { mode: h.mode!, hosts: [...h.hosts] };
};

// The lists of the form's fields: one host per line (the server checks
// and cleans them).
export const homeServersFrom = (mode: HomeServers["mode"], hosts: string): HomeServers => ({
  mode,
  hosts: hosts
    .split("\n")
    .map((h) => h.trim())
    .filter(Boolean),
});
