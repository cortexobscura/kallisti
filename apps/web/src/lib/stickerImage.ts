import { encodeEmoji, loadEmojiImage, renderEmoji } from "./emojiCanvas";
import { isAnimated, keepsOriginal } from "./emojiResize";
import { STICKER_SIZE } from "./stickers";

// A sticker file as it is uploaded without a preview (many at once):
// animations and small images as they are, larger still images scaled to
// fit the sticker size (browser only, as emojiCanvas).
export const fitSticker = async (file: File): Promise<File> => {
  if (isAnimated(new Uint8Array(await file.arrayBuffer()))) return file;
  const { image, original } = await loadEmojiImage(file, STICKER_SIZE);
  if (keepsOriginal({ ...original, animated: false, cropping: false, max: STICKER_SIZE })) return file;
  const blob = await encodeEmoji(renderEmoji(image, null, STICKER_SIZE));
  return new File([blob], file.name.replace(/\.[^.]*$/, "") + (blob.type === "image/webp" ? ".webp" : ".png"), { type: blob.type });
};
