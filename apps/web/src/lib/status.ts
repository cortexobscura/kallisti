import type { ServerState } from "@kallisti/protocol";
import { refusedForGood } from "./reconnect";

// Null while the connection is healthy, otherwise a message for the user.
// Close reasons meant for people (a guest refused on another site, the
// desktop app's own bridge): the others are technical, or another
// server's text.
const shown = (reason: string, localBridge: boolean) => refusedForGood(reason) || (localBridge && reason.startsWith("cannot reach "));

// closeReason: the reason for closing the socket, if one was given;
// localBridge: the desktop app's own bridge closed it (a Mumble server).
export const connectionError = (state: ServerState, closed: boolean, closeReason = "", localBridge = false): string | null => {
  if (state.status === "rejected") return state.error || "The server rejected the connection.";
  if (state.status === "kicked") return state.error || "You were removed from the server.";
  if (!closed) return null;
  return shown(closeReason, localBridge) ? closeReason : "Lost the connection to the server.";
};

// The Kallisti server serves the web app and the bridge (at /ws) from one
// origin. Pages served over https must use wss, or the browser blocks it.
export const proxyUrl = ({ protocol, host }: { protocol: string; host: string }) =>
  `${protocol === "https:" ? "wss" : "ws"}://${host}/ws`;
