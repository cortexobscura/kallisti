import { inVoice, isLobby, joinChannel, setSelfState, type ServerState } from "@kallisti/protocol";
import type { RemoteVoice } from "../components/Client.svelte";
import type { ConnectionInfo } from "./connectionInfo.svelte";
import type { Place } from "./place";
import { voiceJoined } from "./voice";

// Voice across servers (several servers in one client): the clients of all
// connected servers stay mounted (one shown), so voice goes on while another
// is on screen. One voice channel at a time: the server that joined last
// stays, the others leave (once: leaving takes the server's answer) and do
// not rejoin after a reconnect. The others show its voice bar; push-to-talk
// acts on it.

// What it needs of a server's session and connection.
type VoiceConnection = {
  state: ServerState;
  text: { recording?: readonly number[] | null };
  info: ConnectionInfo;
  capture: { setPttDown(down: boolean): void; transmitting: boolean } | null;
  closed: boolean;
  send(frame: Uint8Array): void;
};
export type VoiceView = {
  id: string; // "": the home server, else an added server's origin
  session: { lost: boolean; error: string | null; connection: VoiceConnection | null; place: Place | null; keepPlace(p: Place): void };
};

const me = (s: ServerState) => (s.session === null ? undefined : s.users[s.session]);

export const createVoiceAcross = (deps: {
  views: () => readonly VoiceView[];
  // A server's name (the rail's) and its key (its site), for the bar.
  name: (id: string) => string;
  key: (id: string) => string;
  show: (id: string) => void;
  shown: () => string | null;
}) => {
  // A lost or closed connection's voice is not one (nothing goes through it).
  const inVoiceOn = (s: VoiceView["session"]) =>
    !s.lost && !s.error && !s.connection?.closed && s.connection?.state.status === "synced" && inVoice(s.connection.state);
  const voiceId = $derived(deps.views().find((v) => inVoiceOn(v.session))?.id ?? null);
  const leaveVoiceOn = (id: string) => {
    const c = deps.views().find((v) => v.id === id)?.session.connection;
    const self = c && me(c.state);
    if (c && self && isLobby(c.state)) c.send(joinChannel(self.session, 0));
  };
  let before: string[] = [];
  const dispose = $effect.root(() => {
    $effect(() => {
      const views = deps.views();
      const ids = views.filter((v) => inVoiceOn(v.session)).map((v) => v.id);
      const stays = voiceJoined(before, ids);
      before = ids;
      if (stays === null) return;
      for (const v of views.filter((x) => x.id !== stays)) {
        if (ids.includes(v.id)) leaveVoiceOn(v.id);
        const place = v.session.place;
        if (place?.voice != null) v.session.keepPlace({ ...place, voice: null });
      }
    });
  });
  return {
    get voiceId() {
      return voiceId;
    },
    // The voice bar of the server in voice, for another's screen (null:
    // none, or voice is there).
    remoteVoiceFor: (id: string): RemoteVoice | null => {
      const v = voiceId === null || voiceId === id ? null : deps.views().find((x) => x.id === voiceId);
      const c = v?.session.connection;
      const self = c && me(c.state);
      if (!v || !c || !self) return null;
      const target = v.id;
      return {
        key: deps.key(target),
        server: deps.name(target),
        channel: c.state.channels[self.channelId]?.name ?? "",
        recording: (c.text.recording ?? []).includes(self.channelId),
        selfMute: self.selfMute,
        selfDeaf: self.selfDeaf,
        info: c.info,
        setSelf: (st) => c.send(setSelfState(self.session, st)),
        ptt: (down) => c.capture?.setPttDown(down),
        talking: c.capture?.transmitting ?? false,
        leave: () => leaveVoiceOn(target),
        show: () => deps.show(target),
      };
    },
    // Push-to-talk acts on the server in voice, else on the one shown.
    pttOn: (id: string) => (voiceId ?? deps.shown() ?? "") === id,
    dispose,
  };
};
