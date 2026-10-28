// @vitest-environment jsdom
import { initialState, joinChannel, type ServerState } from "@kallisti/protocol";
import { flushSync } from "svelte";
import { describe, expect, it } from "vitest";
import { createVoiceAcross, type VoiceView } from "../src/lib/voiceAcross.svelte";

// A server where we are session 1 in channel (0: the lobby; not in voice).
const server = (id: string, channel: number) => {
  let state = $state.raw<ServerState>({
    ...initialState,
    status: "synced",
    session: 1,
    rootPermissions: 0, // a lobby (Kallisti)
    channels: { 0: { id: 0, name: "Root", parent: null, description: "", position: 0, temporary: false, links: [] }, 5: { id: 5, name: "General", parent: 0, description: "", position: 0, temporary: false, links: [] } } as unknown as ServerState["channels"],
    users: { 1: { session: 1, name: "me", channelId: channel, selfMute: false, selfDeaf: false } } as unknown as ServerState["users"],
  });
  const sent: Uint8Array[] = [];
  let place = $state<{ voice: number | null } | null>(channel ? { voice: channel } : null);
  const connection = {
    get state() {
      return state;
    },
    text: { recording: [5] },
    info: {} as never,
    capture: null,
    closed: false,
    send: (m: Uint8Array) => void sent.push(m),
  };
  const view: VoiceView = {
    id,
    session: {
      lost: false,
      error: null,
      connection,
      get place() {
        return place as never;
      },
      keepPlace: (p) => void (place = p),
    },
  };
  return {
    view,
    sent,
    move: (to: number) => (state = { ...state, users: { 1: { ...state.users[1]!, channelId: to } } }),
    get place() {
      return place;
    },
  };
};

const setup = (...servers: ReturnType<typeof server>[]) => {
  let shown = $state<string | null>(null);
  const shows: string[] = [];
  const v = createVoiceAcross({
    views: () => servers.map((s) => s.view),
    name: (id) => (id ? `server ${id}` : "home"),
    key: (id) => id || "home",
    show: (id) => void shows.push(id),
    shown: () => shown,
  });
  flushSync();
  return { v, shows, show: (id: string | null) => (shown = id) };
};

describe("voice across servers", () => {
  it("is on the server in voice; the others show its bar", () => {
    const home = server("", 5);
    const b = server("b", 0);
    const { v, shows } = setup(home, b);
    expect(v.voiceId).toBe("");
    expect(v.remoteVoiceFor("")).toBeNull();
    const bar = v.remoteVoiceFor("b")!;
    expect([bar.server, bar.channel, bar.recording, bar.key]).toEqual(["home", "General", true, "home"]);
    bar.show();
    expect(shows).toEqual([""]);
    v.dispose();
  });

  it("joining voice elsewhere leaves the other, and it does not rejoin", () => {
    const home = server("", 5);
    const b = server("b", 0);
    const { v } = setup(home, b);
    b.move(5);
    flushSync();
    expect(home.sent).toEqual([joinChannel(1, 0)]);
    expect(home.place?.voice).toBeNull();
    // Until home's answer, both are in voice: no second leave.
    b.move(5);
    flushSync();
    expect(home.sent).toHaveLength(1);
    v.dispose();
  });

  it("push-to-talk acts on the server in voice, else on the one shown", () => {
    const home = server("", 0);
    const b = server("b", 0);
    const { v, show } = setup(home, b);
    expect([v.pttOn(""), v.pttOn("b")]).toEqual([true, false]);
    show("b");
    flushSync();
    expect([v.pttOn(""), v.pttOn("b")]).toEqual([false, true]);
    home.move(5);
    flushSync();
    expect([v.pttOn(""), v.pttOn("b")]).toEqual([true, false]);
    v.dispose();
  });
});
