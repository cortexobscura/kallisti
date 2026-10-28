import { describe, expect, it } from "vitest";
import { serverPushScope } from "../src/lib/push";
import { clickTarget, noticeOf, serverOfScope } from "../src/lib/pushNotice";

const HOME = "https://chat.example";
const B = "https://b.example";

describe("push for added servers", () => {
  it("has a service worker scope per server, and reads the server back", () => {
    expect(serverPushScope(B, HOME)).not.toMatch(/%2f|%5c/i); // browsers refuse such scopes
    expect(serverPushScope(B, HOME).startsWith(`${HOME}/push/`)).toBe(true);
    expect(serverOfScope(`${HOME}/`, HOME)).toBe(HOME);
    expect(serverOfScope(serverPushScope(B, HOME), HOME)).toBe(B); // the page's scope, the worker's reading
    // Anything else counts as the home server.
    expect(serverOfScope(`${HOME}/push/not-base64!/`, HOME)).toBe(HOME);
    expect(serverOfScope(`${HOME}/push/${btoa("not an origin")}/`, HOME)).toBe(HOME);
  });

  it("links and icons of a notice come from its own server only", () => {
    const n = noticeOf({ title: "t", url: `${B}/channels/1/2`, icon: `${B}/api/logo/192.png?v=1` }, B);
    expect(n.options.data.url).toBe(`${B}/channels/1/2`);
    expect(n.options.icon).toBe(`${B}/api/logo/192.png?v=1`);
    expect(clickTarget(`${HOME}/x`, B)).toBe(`${B}/`);
  });
});

describe("the service worker", () => {
  it("stays one file: only it imports pushNotice (a shared module would become a chunk it imports)", async () => {
    const { readdirSync, readFileSync, statSync } = await import("node:fs");
    const { join } = await import("node:path");
    const files = (dir: string): string[] =>
      readdirSync(dir).flatMap((f) => (statSync(join(dir, f)).isDirectory() ? files(join(dir, f)) : [join(dir, f)]));
    const src = join(__dirname, "../src");
    const importers = files(src).filter((f) => /\.(ts|svelte)$/.test(f) && /from "\.{1,2}\/(lib\/)?pushNotice"/.test(readFileSync(f, "utf8")));
    expect(importers.map((f) => f.slice(src.length + 1))).toEqual(["sw.ts"]);
  });
});
