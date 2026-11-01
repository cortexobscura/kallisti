// The settings' categories (see SettingsDialog): on desktop a menu beside
// the open one, on phones a list where each opens as its own page.

export type SectionId = "profile" | "account" | "voice" | "notifications" | "appearance";
export type Section = { id: SectionId; title: string; group: "user" | "app" };

const ALL: Section[] = [
  { id: "profile", title: "Profile", group: "user" },
  { id: "account", title: "Account", group: "user" },
  { id: "voice", title: "Voice & video", group: "app" },
  { id: "notifications", title: "Notifications", group: "app" },
  { id: "appearance", title: "Appearance", group: "app" },
];

export const GROUP_TITLES = { user: "User settings", app: "App settings" } as const;

// The categories there is something for (no account: no Account).
export const sectionsFor = (has: { profile: boolean; account: boolean }): Section[] =>
  ALL.filter((s) => (s.id === "profile" ? has.profile : s.id === "account" ? has.account : true));

// Where the settings open: the asked category if it is there; else phones
// at the list (null), desktop at the first.
export const startSection = (asked: SectionId | undefined, sections: readonly Section[], phone: boolean): SectionId | null => {
  if (asked && sections.some((s) => s.id === asked)) return asked;
  return phone ? null : (sections[0]?.id ?? null);
};
