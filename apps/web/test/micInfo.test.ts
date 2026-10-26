import { describe, expect, it } from "vitest";
import { micInfo, micInfoRequest } from "../src/lib/voice/micInfo";

describe("micInfo", () => {
  it("says the microphone and the filters that run", () => {
    expect(
      micInfo({ label: "Headset (Bluetooth)", settings: { sampleRate: 16000, channelCount: 1, noiseSuppression: true, echoCancellation: true, autoGainControl: true }, running: "browser", music: false }),
    ).toEqual({ name: "Headset (Bluetooth)", sampleRate: 16000, channels: 1, filters: ["browser noise suppression", "echo cancellation", "automatic gain"] });
    expect(micInfo({ label: "USB mic", settings: { sampleRate: 48000, noiseSuppression: false }, running: "strongest", music: false }).filters).toEqual(["DeepFilterNet"]);
    expect(micInfo({ label: "USB mic", settings: {}, running: "strong", music: false }).filters).toEqual(["RNNoise"]);
    expect(micInfo({ label: "", settings: { channelCount: 2 }, running: null, music: true })).toEqual({ name: "", sampleRate: 0, channels: 2, filters: ["music mode"] });
  });
  it("is cut to what the server takes", () => {
    const info = micInfo({ label: "x".repeat(300), settings: {}, running: null, music: false });
    expect(info.name).toHaveLength(100);
    expect(JSON.parse(micInfoRequest(info))).toEqual({ type: "mic-info", mic: info });
  });
});
