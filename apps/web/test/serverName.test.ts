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
    expect(await info(200, { password: false, name: "Our place" })).toEqual({ password: false, name: "Our place" });
    expect(await info(200, { password: true })).toEqual({ password: true, name: null }); // with Murmur
    expect(await info(404, {})).toEqual({ password: true, name: null }); // an older server
  });
});
