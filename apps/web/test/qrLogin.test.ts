import { describe, expect, it } from "vitest";
import { deviceLabel, parseQrLink, qrLink, qrPath } from "../src/lib/qrLogin";

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
