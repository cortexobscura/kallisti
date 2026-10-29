// Copies text that comes later (an invite link the server makes first).
// Safari copies only inside the click: it takes a promise in a
// ClipboardItem; elsewhere the text is written once it is there.
type Item = new (parts: Record<string, Promise<Blob>>) => unknown;
type Clipboard = { write?: (items: never[]) => Promise<void>; writeText: (text: string) => Promise<void> };

export const copyLater = async (text: Promise<string>, clipboard: Clipboard | undefined, item: Item | undefined) => {
  if (!clipboard) throw new Error("no clipboard");
  if (item && clipboard.write) {
    const blob = text.then((t) => new Blob([t], { type: "text/plain" }));
    return clipboard.write([new item({ "text/plain": blob })] as never[]);
  }
  return clipboard.writeText(await text);
};
