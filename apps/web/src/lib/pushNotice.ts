// The service worker's side of push notifications (see sw.ts): the server
// sends a short text and a link to the message (see server/internal/push).
// Only the service worker imports this (it must stay one file).

const ICON = "/icons/icon-192.png";

// A link opens on this server only.
export const clickTarget = (url: unknown, origin: string): string => {
  try {
    const u = new URL(typeof url === "string" ? url : "/", origin);
    if (u.origin === origin) return u.href;
  } catch {
    // Not a URL.
  }
  return `${origin}/`;
};

// The logo's icon, from this server only (else Kallisti's).
const iconOf = (icon: unknown, origin: string) => {
  try {
    const u = new URL(typeof icon === "string" ? icon : "", origin);
    if (u.origin === origin && u.pathname.startsWith("/api/logo/")) return u.href;
  } catch {
    // Not a URL.
  }
  return ICON;
};

// The server's name and logo come with the push (none: Kallisti's).
export const noticeOf = (data: unknown, origin: string) => {
  const d = data && typeof data === "object" ? (data as Record<string, unknown>) : {};
  const text = (v: unknown) => (typeof v === "string" && v ? v : undefined);
  return {
    title: text(d.title) ?? "Kallisti",
    options: {
      body: text(d.body) ?? "New activity",
      tag: text(d.tag),
      icon: iconOf(d.icon, origin),
      badge: ICON,
      data: { url: clickTarget(d.url, origin) },
    },
  };
};

// The app's window to use for a click: the focused one, else any (null:
// open a new one).
export const windowFor = <W extends { url: string; focused: boolean }>(windows: readonly W[], origin: string): W | null => {
  const ours = windows.filter((w) => w.url.startsWith(`${origin}/`));
  return ours.find((w) => w.focused) ?? ours[0] ?? null;
};

// Push for added servers (several servers in one client): one service
// worker registration each (a subscription is per registration, and each
// server has its own key), at <home>/push/<origin>/ (see lib/push).

// The server a registration is for (its scope, see lib/push
// serverPushScope; anything else: home).
export const serverOfScope = (scope: string, home: string): string => {
  const prefix = `${home}/push/`;
  if (!scope.startsWith(prefix) || !scope.endsWith("/")) return home;
  try {
    const server = atob(scope.slice(prefix.length, -1).replaceAll("-", "+").replaceAll("_", "/"));
    return new URL(server).origin === server ? server : home;
  } catch {
    return home;
  }
};
