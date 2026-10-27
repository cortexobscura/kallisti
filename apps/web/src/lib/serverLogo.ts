// The server's logo (admins upload it): icons by size at
// /api/logo/<size>.png?v=<version>; without one, Kallisti's.
const defaults: Record<number, string> = { 32: "/icons/favicon-32.png", 180: "/icons/apple-touch-icon.png", 192: "/icons/icon-192.png" };

export const iconHref = (base: string, size: 32 | 180 | 192, version: string | null) =>
  version ? `${base}/api/logo/${size}.png?v=${encodeURIComponent(version)}` : defaults[size]!;

// The tab icon and the home screen icon follow the logo (the server puts
// it in the page too; this is for changes while it is open). null:
// Kallisti's.
export const showIcons = (doc: Document, base: string, version: string | null) => {
  doc.querySelector<HTMLLinkElement>('link[rel="icon"]')?.setAttribute("href", iconHref(base, 32, version));
  doc.querySelector<HTMLLinkElement>('link[rel="apple-touch-icon"]')?.setAttribute("href", iconHref(base, 180, version));
};
