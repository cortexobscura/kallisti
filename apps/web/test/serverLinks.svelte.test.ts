// @vitest-environment jsdom
import { flushSync } from "svelte";
import { describe, expect, it } from "vitest";
import { createServerLinks } from "../src/lib/serverLinks.svelte";

const HOME = "https://home.example";
const B = "https://b.example";
const C = "https://c.example";
const TOKEN = "abcdefghij0123456789xyz";

const setup = () => {
  let owner = $state<string | null>(`${HOME} alice`);
  const servers = [{ origin: B, name: "al", session: {} }, { origin: C, name: "al", session: null }];
  const shown: string[] = [];
  const adds: unknown[] = [];
  const links = createServerLinks({
    servers: () => servers,
    home: () => HOME,
    owner: () => owner,
    select: (id) => void shown.push(id),
    add: (target) => void adds.push(target),
  });
  flushSync();
  return { links, shown, adds, login: (o: string | null) => (owner = o) };
};

describe("links from elsewhere", () => {
  it("an invite: home, a server here, its login first, or added first", () => {
    const { links, shown, adds } = setup();
    expect(links.openInvite(`${HOME}/invite/${TOKEN}`)).toBe(true);
    expect(links.openInvite(`${B}/invite/${TOKEN}`)).toBe(true);
    expect(links.openInvite(`${C}/invite/${TOKEN}`)).toBe(true);
    expect(links.openInvite(`https://new.example/invite/${TOKEN}`)).toBe(true);
    expect(shown).toEqual(["", B, C]);
    expect(adds).toEqual([{ server: { origin: C, name: "al" } }, { address: "https://new.example" }]);
    expect([links.linkFor(""), links.linkFor(B)]).toEqual([`${HOME}/invite/${TOKEN}`, `${B}/invite/${TOKEN}`]);
    links.used(B);
    expect(links.linkFor(B)).toBeNull();
    expect(links.openInvite("https://x.example/not-an-invite")).toBe(false);
    links.dispose();
  });

  it("a guest adds no servers: the link opens as a page", () => {
    const { links, login } = setup();
    login(null);
    flushSync();
    expect(links.openInvite(`https://new.example/invite/${TOKEN}`)).toBe(false);
    links.dispose();
  });

  it("a message link (a push notice) opens on a server here, nothing else", () => {
    const { links, shown } = setup();
    links.openMessageLink(`${B}/channels/1/2`);
    links.openMessageLink(`${B}/invite/${TOKEN}`);
    links.openMessageLink(`${C}/channels/1/2`); // no session here
    expect(shown).toEqual([B]);
    expect(links.linkFor(B)).toBe(`${B}/channels/1/2`);
    links.dispose();
  });

  it("another home login: none of the last one's links", () => {
    const { links, login } = setup();
    links.openInvite(`${B}/invite/${TOKEN}`);
    login(`${HOME} bob`);
    flushSync();
    expect(links.linkFor(B)).toBeNull();
    links.dispose();
  });
});
