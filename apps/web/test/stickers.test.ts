import { describe, expect, it } from "vitest";
import { parseStickers, stickerForm, stickerGroups, stickerSearch, stickerText, type Sticker } from "../src/lib/stickers";

const wave: Sticker = { id: 1, name: "wave", description: "Hello there", emoji: "👋", pack: "Greetings", url: "/api/stickers/1" };
const cat: Sticker = { id: 2, name: "cat", url: "/api/stickers/2" };
const fox: Sticker = { id: 3, name: "fox_hug", description: "A hug", emoji: "🤗", pack: "neofox", url: "/api/stickers/3" };

describe("parseStickers", () => {
  it("keeps good entries, drops broken ones and foreign images", () => {
    const got = parseStickers([wave, cat, { id: 4, name: "x", url: "/api/stickers/4" }, { id: 5, name: "evil", url: "https://example.org/a.png" }, { name: "noid", url: "/api/stickers/6" }]);
    expect(got).toEqual([wave, cat]);
    expect(parseStickers(undefined)).toBeNull();
  });
});

describe("stickerGroups", () => {
  it("groups by pack, stickers without one first, in the list's order", () => {
    expect(stickerGroups([cat, wave, fox])).toEqual([
      { pack: "", stickers: [cat] },
      { pack: "Greetings", stickers: [wave] },
      { pack: "neofox", stickers: [fox] },
    ]);
  });
});

describe("stickerSearch", () => {
  it("finds by name, description, pack and emoji, ignoring case", () => {
    const all = [wave, cat, fox];
    expect(stickerSearch(all, "")).toEqual(all);
    expect(stickerSearch(all, "HELLO")).toEqual([wave]);
    expect(stickerSearch(all, "hug")).toEqual([fox]);
    expect(stickerSearch(all, "neo")).toEqual([fox]);
    expect(stickerSearch(all, "👋")).toEqual([wave]);
    expect(stickerSearch(all, "dog")).toEqual([]);
  });
});

describe("stickerText", () => {
  it("is what text-only places show", () => {
    expect(stickerText("wave")).toBe("[sticker: wave]");
  });
});

describe("stickerForm", () => {
  it("checks the fields as the server does", () => {
    expect(stickerForm({ name: "wave", description: "", emoji: "", pack: "" }, [cat])).toBeNull();
    expect(stickerForm({ name: "Wave", description: "", emoji: "", pack: "" }, [])).toMatch(/lower case/);
    expect(stickerForm({ name: "cat", description: "", emoji: "", pack: "" }, [cat])).toMatch(/exists/);
    expect(stickerForm({ name: "cat", description: "", emoji: "", pack: "" }, [cat], 2)).toBeNull(); // itself
    expect(stickerForm({ name: "wave", description: "x".repeat(101), emoji: "", pack: "" }, [])).toMatch(/100/);
    expect(stickerForm({ name: "wave", description: "", emoji: "", pack: "x".repeat(33) }, [])).toMatch(/32/);
    expect(stickerForm({ name: "wave", description: "", emoji: "abc", pack: "" }, [])).toMatch(/emoji/);
    expect(stickerForm({ name: "wave", description: "", emoji: "👋", pack: "" }, [])).toBeNull();
  });
});
