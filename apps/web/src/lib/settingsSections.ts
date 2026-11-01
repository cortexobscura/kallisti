import { startCategory, type Category } from "./dialogCategories";

// The user settings' categories (see SettingsDialog, lib/categories).

export type SectionId = "profile" | "account" | "voice" | "notifications" | "appearance";
export type Section = Category<SectionId, "user" | "app">;

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

export const startSection = startCategory<SectionId>;
