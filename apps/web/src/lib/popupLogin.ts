import type { FediResult } from "./fedi";

// The popup in which a Kallisti page adds this server (see Connect): the
// home server that asked (?kallisti-connect=, kept in the tab as mark over
// the login's pages: kept false when the tab keeps nothing), and where it
// is: "start" the login, "deliver" the token, an error, or null (the name
// step: new, or again after a reload, resumed). home null: no popup.
export type ConnectPhase = "start" | "deliver" | { error: string } | null;

export const popupLogin = (s: { asked: string | null; mark: string | null; back: FediResult | null; kept: boolean }) => {
  const resumed = !s.asked && !s.back && !!s.mark;
  const back = (s.back && ["connect", "register", "error"].includes(s.back.kind)) || resumed ? s.mark : null;
  const home = s.asked ?? back;
  const phase: ConnectPhase = !home
    ? null
    : !s.kept
      ? { error: "This browser keeps no data for this window. Log in with a password there instead." }
      : resumed || s.back?.kind === "register"
        ? null
        : !s.back
          ? "start"
          : s.back.kind === "connect"
            ? "deliver"
            : s.back.kind === "error"
              ? { error: s.back.text }
              : null;
  return { home, resumed, phase };
};
