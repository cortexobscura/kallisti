import { describe, expect, it } from "vitest";
import { copyLater } from "../src/lib/clipboard";

describe("copyLater", () => {
  it("hands the clipboard a promise where it takes one (Safari: inside the click)", async () => {
    const written: Blob[] = [];
    class Item {
      constructor(readonly parts: Record<string, Promise<Blob>>) {}
    }
    const clipboard = {
      write: async (items: Item[]) => void written.push(await items[0]!.parts["text/plain"]!),
      writeText: async () => {
        throw new Error("not used");
      },
    };
    await copyLater(Promise.resolve("https://x/invite/abc"), clipboard, Item);
    expect(await written[0]!.text()).toBe("https://x/invite/abc");
  });

  it("writes at once, before the text comes (still inside the click)", async () => {
    let wrote = false;
    let give: (t: string) => void = () => {};
    const text = new Promise<string>((r) => (give = r));
    class Item {
      constructor(readonly parts: Record<string, Promise<Blob>>) {}
    }
    const done = copyLater(text, { write: async () => void (wrote = true), writeText: async () => {} }, Item);
    expect(wrote).toBe(true);
    give("https://x/invite/abc");
    await done;
  });

  it("fails when the text fails after the write", async () => {
    class Item {
      constructor(readonly parts: Record<string, Promise<Blob>>) {}
    }
    await expect(copyLater(Promise.reject(new Error("refused")), { write: async () => {}, writeText: async () => {} }, Item)).rejects.toThrow("refused");
  });

  it("else writes the text once it is there", async () => {
    let text = "";
    await copyLater(Promise.resolve("https://x/voice/1"), { writeText: async (t: string) => void (text = t) }, undefined);
    expect(text).toBe("https://x/voice/1");
  });

  it("fails when the text does not come", async () => {
    await expect(copyLater(Promise.reject(new Error("no")), { writeText: async () => {} }, undefined)).rejects.toThrow("no");
  });
});
