import { describe, expect, it } from "vitest";
import { channelLink, linkPlan, messageLink, parseMessageLink, voiceLink } from "../src/lib/messageLink";

describe("message links", () => {
  it("make a link to a message on the server", () => {
    expect(messageLink("https://chat.example", 4, 1234)).toBe("https://chat.example/channels/4/1234");
  });

  it("read a link back, only for this server", () => {
    expect(parseMessageLink("https://chat.example/channels/4/1234", "https://chat.example")).toEqual({ chat: 4, message: 1234 });
    expect(parseMessageLink("/channels/4/1234", "https://chat.example")).toEqual({ chat: 4, message: 1234 });
    expect(parseMessageLink("https://other.example/channels/4/1234", "https://chat.example")).toBeNull();
    expect(parseMessageLink("https://chat.example/channels/x/1", "https://chat.example")).toBeNull();
    expect(parseMessageLink("https://chat.example/", "https://chat.example")).toBeNull();
  });

  it("make and read a link to a chat", () => {
    expect(channelLink("https://chat.example", 4)).toBe("https://chat.example/channels/4");
    expect(parseMessageLink("https://chat.example/channels/4", "https://chat.example")).toEqual({ chat: 4 });
    expect(parseMessageLink("https://chat.example/channels/4/", "https://chat.example")).toBeNull();
  });
});

describe("voice channel links", () => {
  it("make and read a link to a voice channel (root too)", () => {
    expect(voiceLink("https://chat.example", 5)).toBe("https://chat.example/voice/5");
    expect(parseMessageLink("https://chat.example/voice/5", "https://chat.example")).toEqual({ voice: 5 });
    expect(parseMessageLink("https://chat.example/voice/0", "https://chat.example")).toEqual({ voice: 0 });
    expect(parseMessageLink("https://other.example/voice/5", "https://chat.example")).toBeNull();
    expect(parseMessageLink("https://chat.example/voice/5/1", "https://chat.example")).toBeNull();
  });
});

describe("what Copy link makes", () => {
  const site = "https://chat.example";
  it("an invite for who manages a private channel, else a plain link", () => {
    expect(linkPlan(site, "voice", 5, { access: 5, invites: true })).toEqual({ invite: 5 });
    expect(linkPlan(site, "text", 4, { access: 20, invites: true })).toEqual({ invite: 20 });
    expect(linkPlan(site, "voice", 5, { access: 5, invites: false })).toEqual({ url: "https://chat.example/voice/5" });
    expect(linkPlan(site, "text", 4, { access: null, invites: false })).toEqual({ url: "https://chat.example/channels/4" });
    expect(linkPlan(site, "category", 1, { access: null, invites: false })).toBeNull();
    // A private category: an invite for who manages it, nothing for others.
    expect(linkPlan(site, "category", 1, { access: 20, invites: true })).toEqual({ invite: 20 });
    expect(linkPlan(site, "category", 1, { access: 20, invites: false })).toBeNull();
  });
});
