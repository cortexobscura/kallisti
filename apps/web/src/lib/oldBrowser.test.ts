// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import script from "../../public/old-browser.js?raw";

// public/old-browser.js runs as a plain script before the app: run here in
// a browser with (or without) the features the app needs.
const toSorted = Object.getOwnPropertyDescriptor(Array.prototype, "toSorted")!;

const page = (userAgent: string, modern: boolean) => {
  document.body.innerHTML = '<div id="app"></div>';
  vi.spyOn(navigator, "userAgent", "get").mockReturnValue(userAgent);
  if (!modern) Reflect.deleteProperty(Array.prototype, "toSorted");
  try {
    new Function(script)();
  } finally {
    Object.defineProperty(Array.prototype, "toSorted", toSorted);
  }
  return document.body.textContent;
};

afterEach(() => vi.restoreAllMocks());

const chrome = "Mozilla/5.0 (Linux; Android 7.0) AppleWebKit/537.36 Chrome/52.0 Mobile Safari/537.36";
const webView = "Mozilla/5.0 (Linux; Android 7.0; wv) AppleWebKit/537.36 Version/4.0 Chrome/52.0 Mobile Safari/537.36";

describe("old-browser.js", () => {
  it("leaves the page alone in a browser the app runs in", () => {
    expect(page(chrome, true)).toBe("");
  });

  it("says to update the browser in an old one", () => {
    const text = page(chrome, false);
    expect(text).toContain("too old");
    expect(text).toContain("Update your browser");
  });

  it("keeps the app from mounting below the message", () => {
    page(chrome, false);
    expect(document.getElementById("app")).toBeNull();
  });

  it("says to update Android System WebView in the Android app", () => {
    expect(page(webView, false)).toContain("Android System WebView");
  });
});
