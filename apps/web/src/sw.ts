// The service worker, for push notifications only (see lib/push.ts): no
// fetch handler, so it caches nothing and the app loads as without it.
// Built to /sw.js (see vite.config.ts); it must stay one file.
import { clickTarget, noticeOf, serverOfScope, windowFor } from "./lib/pushNotice";

// The few worker types used here (the WebWorker lib clashes with DOM).
type Waitable = Event & { waitUntil(p: Promise<unknown>): void };
type AppWindow = { url: string; focused: boolean; focus(): Promise<unknown>; postMessage(m: unknown): void };
type Worker = {
  registration: ServiceWorkerRegistration;
  clients: { matchAll(o: { type: "window"; includeUncontrolled: boolean }): Promise<AppWindow[]>; openWindow(url: string): Promise<unknown> };
  skipWaiting(): Promise<void>;
  addEventListener(type: string, f: (e: Waitable) => void): void;
};
const sw = self as unknown as Worker;
// The page's origin (the home server), and the server this registration's
// pushes come from (an added server's: see pushScope).
const origin = location.origin;
const server = serverOfScope(sw.registration.scope, origin);

// A new version takes over at once (it holds no state).
sw.addEventListener("install", () => void sw.skipWaiting());

sw.addEventListener("push", (e) => {
  let data: unknown = null;
  try {
    data = (e as unknown as { data: { json(): unknown } | null }).data?.json();
  } catch {
    // Not JSON: the general text.
  }
  const n = noticeOf(data, server);
  e.waitUntil(sw.registration.showNotification(n.title, n.options));
});

// A click opens the message: in an open window of the app (which jumps
// there; an added server's client takes its own links), else in a new one
// (the server answers message links with the app; the home page opens an
// added server's with ?open=).
sw.addEventListener("notificationclick", (e) => {
  const notification = (e as unknown as { notification: Notification }).notification;
  notification.close();
  const url = clickTarget((notification.data as { url?: unknown } | null)?.url, server);
  e.waitUntil(
    (async () => {
      const win = windowFor(await sw.clients.matchAll({ type: "window", includeUncontrolled: true }), origin);
      if (!win) return sw.clients.openWindow(server === origin ? url : `${origin}/?open=${encodeURIComponent(url)}`);
      await win.focus();
      win.postMessage({ type: "open", url });
    })(),
  );
});
