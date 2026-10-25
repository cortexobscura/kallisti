// The iOS app (WKWebView): WebKit interrupts the audio context when the
// app goes to the background (another app, the screen locked), and voice
// stops: the microphone stays open, but nothing takes its sound or plays.
// The app has the audio background mode: a context resumed once the page
// counts as hidden runs on. Elsewhere contexts are not "interrupted" here.
// Only while in voice (Client.svelte): a running context keeps the app
// awake.

type Doc = {
  visibilityState: DocumentVisibilityState;
  addEventListener: (type: "visibilitychange", listener: () => void) => void;
  removeEventListener: (type: "visibilitychange", listener: () => void) => void;
};
type Ctx = {
  readonly state: string; // "interrupted" is newer than TypeScript's list
  resume: () => Promise<void>;
  addEventListener: (type: "statechange", listener: () => void) => void;
  removeEventListener: (type: "statechange", listener: () => void) => void;
};

// Returns stop. An audio restart: stop, then again with the new context.
export const keepRunningHidden = (doc: Doc, ctx: Ctx): (() => void) => {
  const check = () => {
    if (doc.visibilityState === "hidden" && ctx.state === "interrupted") ctx.resume().catch(() => {});
  };
  // The interruption comes just before the page counts as hidden; later
  // ones (a phone call) while it is hidden.
  doc.addEventListener("visibilitychange", check);
  ctx.addEventListener("statechange", check);
  return () => {
    doc.removeEventListener("visibilitychange", check);
    ctx.removeEventListener("statechange", check);
  };
};
