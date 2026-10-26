import type { NoiseMode } from "./noise";

// What the browser says about the microphone in use, for the admins' view
// of connections (see the server's mic-info): a sample rate of 16 or 8 kHz
// is a Bluetooth headset in its phone mode, which sounds muffled.

export type MicInfo = { name: string; sampleRate: number; channels: number; filters: string[] };

const MAX_NAME = 100; // as the server

type Settings = Pick<MediaTrackSettings, "sampleRate" | "channelCount" | "noiseSuppression" | "echoCancellation" | "autoGainControl">;

export const micInfo = (m: { label: string; settings: Settings; running: NoiseMode | null; music: boolean }): MicInfo => ({
  name: [...m.label].slice(0, MAX_NAME).join(""),
  sampleRate: m.settings.sampleRate ?? 0,
  channels: m.settings.channelCount ?? 0,
  filters: [
    ...(m.settings.noiseSuppression ? ["browser noise suppression"] : []),
    ...(m.settings.echoCancellation ? ["echo cancellation"] : []),
    ...(m.settings.autoGainControl ? ["automatic gain"] : []),
    ...(m.running === "strong" ? ["RNNoise"] : m.running === "strongest" ? ["DeepFilterNet"] : []),
    ...(m.music ? ["music mode"] : []),
  ],
});

export const micInfoRequest = (mic: MicInfo) => JSON.stringify({ type: "mic-info", mic });
