import { describe, expect, it } from "vitest";
import { createAccountsApi } from "../src/lib/accounts";
import { confirmsManage } from "../src/lib/saveFeedback";
import { manageRequest, parseServerEvent } from "../src/lib/textChannels";

const fetchJson = (status: number, body: unknown) =>
  (async () => new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } })) as unknown as typeof fetch;

describe("the server's name", () => {
  it("is set by admins and confirmed by the server's event", () => {
    expect(manageRequest({ serverName: "Our place" })).toBe('{"type":"set-server-name","name":"Our place"}');
    const event = parseServerEvent('{"type":"server-name","name":"Our place"}')!;
    expect(event).toEqual({ type: "server-name", name: "Our place" });
    expect(parseServerEvent('{"type":"server-name","name":3}')).toBeNull();
    expect(confirmsManage({ serverName: "  Our place " })!(event)).toBe(true);
    expect(confirmsManage({ serverName: "Other" })!(event)).toBe(false);
  });

  it("comes with the server info for the login page", async () => {
    const info = (status: number, body: unknown) => createAccountsApi("http://srv", fetchJson(status, body)).serverInfo();
    expect(await info(200, { password: false, name: "Our place" })).toEqual({ password: false, name: "Our place", logo: null });
    expect(await info(200, { password: true })).toEqual({ password: true, name: null, logo: null }); // with Murmur
    expect(await info(404, {})).toEqual({ password: true, name: null, logo: null }); // an older server
  });
});

describe("the login check", () => {
  const api = (status: number | "offline", body: unknown = {}) =>
    createAccountsApi(
      "http://srv",
      (async () => {
        if (status === "offline") throw new TypeError("Failed to fetch");
        return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
      }) as unknown as typeof fetch,
    );
  it("tells an account, no login, and unknown (a server or network error) apart", async () => {
    expect(await api(200, { name: "alice" }).check()).toMatchObject({ name: "alice" });
    expect(await api(401, { error: "not logged in" }).check()).toBeNull();
    expect(await api(200, {}).check()).toBeNull();
    expect(await api(502).check()).toBe("unknown");
    expect(await api(500).check()).toBe("unknown");
    expect(await api("offline").check()).toBe("unknown");
  });
});
