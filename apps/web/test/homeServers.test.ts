import { describe, expect, it } from "vitest";
import { homeServersFrom, parseHomeServers } from "../src/lib/homeServers";
import { confirmsManage } from "../src/lib/saveFeedback";
import { initialTextState, manageRequest, parseServerEvent, reduceText } from "../src/lib/textChannels";

describe("which other servers' accounts log in", () => {
  const homes = { mode: "block" as const, hosts: ["evil.example"] };
  it("parses what the server says, and nothing else", () => {
    expect(parseHomeServers(homes)).toEqual(homes);
    expect(parseHomeServers({ mode: "maybe", hosts: [] })).toBeNull();
    expect(parseHomeServers({ mode: "all", hosts: [3] })).toBeNull();
  });
  it("makes the lists of the form's fields", () => {
    expect(homeServersFrom("allow", " home.example \n\nother.example\n")).toEqual({ mode: "allow", hosts: ["home.example", "other.example"] });
  });
  it("is asked for, saved and confirmed like other settings", () => {
    const s = reduceText(initialTextState, parseServerEvent(JSON.stringify({ type: "home-servers", homes }))!);
    expect(s.homeServers).toEqual(homes);
    expect(JSON.parse(manageRequest({ homeServers: true }))).toEqual({ type: "home-servers" });
    expect(JSON.parse(manageRequest({ setHomeServers: homes }))).toEqual({ type: "set-home-servers", homes });
    expect(confirmsManage({ setHomeServers: homes })!({ type: "home-servers", homes })).toBe(true);
  });
});
