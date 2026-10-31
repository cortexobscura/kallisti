import { describe, expect, it } from "vitest";
import { channelLink, dmLink, inviteTarget, messageLink, parseMessageLink, plainLink, voiceLink } from "../src/lib/messageLink";

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

describe("what Copy link and Copy invite link make", () => {
  const site = "https://chat.example";
  it("a plain link to a text or voice channel (none for a category)", () => {
    expect(plainLink(site, "voice", 5)).toBe("https://chat.example/voice/5");
    expect(plainLink(site, "text", 4)).toBe("https://chat.example/channels/4");
    expect(plainLink(site, "category", 1)).toBeNull();
  });
  it("an invite to the private channel or category, for who manages it", () => {
    expect(inviteTarget({ access: 20, invites: true })).toBe(20);
    expect(inviteTarget({ access: 20, invites: false })).toBeNull();
    expect(inviteTarget({ access: null, invites: true })).toBeNull();
  });
});

describe("direct chat links", () => {
  it("open a direct chat with a registered user of that server", () => {
    expect(dmLink("https://a.example", 7)).toBe("https://a.example/dm/7");
    expect(parseMessageLink("https://a.example/dm/7", "https://a.example")).toEqual({ dm: 7 });
    expect(parseMessageLink("https://b.example/dm/7", "https://a.example")).toBeNull();
    // Only user ids (32 bits): not some other user by overflow.
    expect(parseMessageLink("https://a.example/dm/4294967295", "https://a.example")).toEqual({ dm: 4294967295 });
    expect(parseMessageLink(`https://a.example/dm/4294967296`, "https://a.example")).toBeNull();
    expect(parseMessageLink(`https://a.example/dm/${"9".repeat(400)}`, "https://a.example")).toBeNull();
  });
});
