import type { ServerState } from "@kallisti/protocol";

// Null while the connection is healthy, otherwise a message for the user.
// Close reasons meant for people (a guest refused on another site, the
// desktop app's bridge): the others are technical.
const shownReasons = [/^join as a guest /, /^cannot reach /]; // see reconnect.refusedForGood

// closeReason: the server's reason for closing the socket, if it gave one.
export const connectionError = (state: ServerState, closed: boolean, closeReason = ""): string | null => {
  if (state.status === "rejected") return state.error || "The server rejected the connection.";
  if (state.status === "kicked") return state.error || "You were removed from the server.";
  if (!closed) return null;
  return shownReasons.some((r) => r.test(closeReason)) ? closeReason : "Lost the connection to the server.";
};

// The Kallisti server serves the web app and the bridge (at /ws) from one
// origin. Pages served over https must use wss, or the browser blocks it.
export const proxyUrl = ({ protocol, host }: { protocol: string; host: string }) =>
  `${protocol === "https:" ? "wss" : "ws"}://${host}/ws`;
