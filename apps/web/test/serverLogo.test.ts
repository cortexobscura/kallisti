import { describe, expect, it } from "vitest";
import { createAccountsApi } from "../src/lib/accounts";
import { confirmsManage } from "../src/lib/saveFeedback";
import { iconHref } from "../src/lib/serverLogo";
import { initialTextState, logoTokenRequest, manageRequest, parseServerEvent, reduceText } from "../src/lib/textChannels";

const fetchJson = (body: unknown) =>
  (async () => new Response(JSON.stringify(body), { status: 200, headers: { "Content-Type": "application/json" } })) as unknown as typeof fetch;

describe("the server's logo", () => {
  it("comes with ready and changes later", () => {
    expect(initialTextState.logo).toBeUndefined(); // the page's icons stay until then
    const s = reduceText(initialTextState, parseServerEvent('{"type":"ready","channels":[],"logo":"abc"}')!);
    expect(s.logo).toBe("abc");
    expect(reduceText(s, parseServerEvent('{"type":"logo","version":"def"}')!).logo).toBe("def");
    expect(reduceText(s, parseServerEvent('{"type":"logo","version":""}')!).logo).toBeNull();
    expect(reduceText(s, parseServerEvent('{"type":"ready","channels":[]}')!).logo).toBeNull();
    expect(parseServerEvent('{"type":"logo","version":3}')).toBeNull();
  });

  it("is uploaded with a token and removed by admins", () => {
    expect(logoTokenRequest()).toBe('{"type":"logo-token"}');
    expect(manageRequest({ removeLogo: true })).toBe('{"type":"remove-logo"}');
    const confirms = confirmsManage({ removeLogo: true })!;
    expect(confirms(parseServerEvent('{"type":"logo","version":""}')!)).toBe(true);
    expect(confirms(parseServerEvent('{"type":"logo","version":"abc"}')!)).toBe(false);
  });

  it("gives the icons' addresses, Kallisti's without a logo", () => {
    expect(iconHref("https://chat.example", 32, "abc")).toBe("https://chat.example/api/logo/32.png?v=abc");
    expect(iconHref("", 180, "abc")).toBe("/api/logo/180.png?v=abc");
    expect(iconHref("", 32, null)).toBe("/icons/favicon-32.png");
    expect(iconHref("", 180, null)).toBe("/icons/apple-touch-icon.png");
  });

  it("comes with the server info for the login page", async () => {
    expect((await createAccountsApi("http://srv", fetchJson({ password: false, name: "x", logo: "abc" })).serverInfo()).logo).toBe("abc");
    expect((await createAccountsApi("http://srv", fetchJson({ password: false })).serverInfo()).logo).toBeNull();
  });
});
