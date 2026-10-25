// The stream view's height, as a share of the window's (it scales with
// the window), dragged by its lower edge and saved. Limits: chat and
// stream both stay usable.
export const STAGE_MIN = 0.2;
export const STAGE_MAX = 0.8;
const STEP = 0.05; // arrow keys

const clamp = (share: number) => Math.min(STAGE_MAX, Math.max(STAGE_MIN, share));

// Not dragged yet (null): the default, smaller on phones.
export const shareOf = (saved: number | null, phone: boolean) => saved ?? (phone ? 0.4 : 0.55);

export const parseShare = (raw: string | null): number | null => {
  try {
    const n: unknown = JSON.parse(raw ?? "null");
    return typeof n === "number" && Number.isFinite(n) ? clamp(n) : null;
  } catch {
    return null;
  }
};

// The share while the edge is dragged from startY to y (window pixels).
export const dragShare = (start: number, startY: number, y: number, windowHeight: number) => clamp(start + (y - startY) / windowHeight);

// null: not a key for this.
export const keyShare = (share: number, key: string): number | null => {
  if (key === "ArrowUp") return clamp(share - STEP);
  if (key === "ArrowDown") return clamp(share + STEP);
  if (key === "Home") return STAGE_MIN;
  if (key === "End") return STAGE_MAX;
  return null;
};
