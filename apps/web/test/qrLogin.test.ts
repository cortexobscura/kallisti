import { describe, expect, it, vi } from "vitest";
import type { QrState } from "../src/lib/accounts";
import { deviceLabel, parseQrLink, pollState, qrLink, qrPath } from "../src/lib/qrLogin";

const token = "ab".repeat(32);

describe("QR login links", () => {
  it("are the server's address with the token", () => {
    expect(qrLink("https://chat.example", token)).toBe(`https://chat.example/qr/${token}`);
    expect(parseQrLink(`https://chat.example/qr/${token}`)).toEqual({ origin: "https://chat.example", token });
  });

  it("take only https (http on this computer) and a whole token", () => {
    expect(parseQrLink(`http://chat.example/qr/${token}`)).toBeNull();
    expect(parseQrLink(`http://localhost:5173/qr/${token}`)).toEqual({ origin: "http://localhost:5173", token });
    expect(parseQrLink("https://chat.example/qr/abc")).toBeNull();
    expect(parseQrLink(`https://chat.example/invite/${token}`)).toBeNull();
    expect(parseQrLink("not a url")).toBeNull();
  });
});

describe("deviceLabel", () => {
  it("names the browser and the system", () => {
    expect(deviceLabel("Mozilla/5.0 (X11; Linux x86_64; rv:150.0) Gecko/20100101 Firefox/150.0", false)).toBe("Firefox on Linux");
    expect(deviceLabel("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0 Safari/537.36 Edg/140.0", false)).toBe(
      "Edge on Windows",
    );
    expect(deviceLabel("Mozilla/5.0 (iPhone; CPU iPhone OS 27_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/27.0 Mobile/15E148 Safari/604.1", false)).toBe(
      "Safari on iPhone",
    );
    expect(deviceLabel("Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0 Mobile Safari/537.36", false)).toBe(
      "Chrome on Android",
    );
    expect(deviceLabel("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko)", true)).toBe("Kallisti app on macOS");
    expect(deviceLabel("", false)).toBe("A browser");
    // iPadOS says "Macintosh"; its touch screen tells.
    const ipad = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/27.0 Safari/605.1.15";
    expect(deviceLabel(ipad, false, 5)).toBe("Safari on iPad");
    expect(deviceLabel(ipad, false, 0)).toBe("Safari on macOS");
  });
});

describe("qrPath", () => {
  it("draws the dark cells as one path, with a quiet border", () => {
    const { size, path } = qrPath("https://chat.example/qr/" + token);
    expect(size).toBeGreaterThan(29); // a long link: version 4 and up, plus the border
    expect(path.startsWith("M")).toBe(true);
    expect(path).not.toContain("M0 0"); // the border is light
  });
});

describe("pollState", () => {
  it("asks until told to stop; a failed ask is asked again; stop ends it", async () => {
    vi.useFakeTimers();
    const answers: (QrState | Error)[] = [{ state: "waiting" }, new Error("lost"), { state: "scanned" }, { state: "done" }];
    const wait = vi.fn(async () => {
      const a = answers.shift()!;
      if (a instanceof Error) throw a;
      return a;
    });
    const seen: string[] = [];
    pollState(wait, (s) => (seen.push(s.state), s.state !== "done"), 100);
    await vi.advanceTimersByTimeAsync(1000);
    expect(seen).toEqual(["waiting", "scanned", "done"]);
    expect(wait).toHaveBeenCalledTimes(4);

    const again = vi.fn(async (): Promise<QrState> => ({ state: "waiting" }));
    const stop = pollState(again, () => true, 100);
    await vi.advanceTimersByTimeAsync(250);
    stop();
    await vi.advanceTimersByTimeAsync(1000);
    expect(again).toHaveBeenCalledTimes(2);

    // No answers (no network): expired at the deadline.
    const states: string[] = [];
    pollState(() => Promise.reject(new Error("offline")), (s) => (states.push(s.state), true), 100, Date.now() + 350);
    await vi.advanceTimersByTimeAsync(1000);
    expect(states).toEqual(["expired"]);

    // A request that never answers: expired at the deadline too.
    const hung: string[] = [];
    pollState(() => new Promise<QrState>(() => {}), (s) => (hung.push(s.state), true), 100, Date.now() + 350);
    await vi.advanceTimersByTimeAsync(1000);
    expect(hung).toEqual(["expired"]);
    vi.useRealTimers();
  });
});
