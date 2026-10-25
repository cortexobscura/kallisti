// Several pictures of a message show as a grid, as in Discord (see
// ChatPane), and the viewer goes through them (see Lightbox).

// A cell's spans in a grid of 6 columns.
export type Cell = { cols: number; rows: number };

// Rows of three; a count that does not fill them starts with a row of one
// or two. Three are one big and two small, four are two by two.
export const gridCells = (n: number): Cell[] => {
  if (n === 3) return [{ cols: 4, rows: 2 }, { cols: 2, rows: 1 }, { cols: 2, rows: 1 }];
  const first = n === 4 ? 2 : n % 3;
  const sizes = n === 4 ? [2, 2] : [...(first ? [first] : []), ...Array<number>(Math.floor(n / 3)).fill(3)];
  return sizes.flatMap((k) => Array.from({ length: k }, () => ({ cols: 6 / k, rows: 1 })));
};

// The pictures for the grid (two or more) and the other files, in order.
export const splitFiles = <U>(uploads: readonly U[], viewOf: (u: U) => string) => {
  const grid = uploads.filter((u) => viewOf(u) === "image");
  return grid.length < 2 ? { grid: [], rest: [...uploads] } : { grid, rest: uploads.filter((u) => viewOf(u) !== "image") };
};

// The viewer's next picture (by=1) or the one before; it stays at the ends.
export const step = (index: number, by: number, count: number) => Math.min(count - 1, Math.max(0, index + by));

// A sideways swipe of at least SWIPE_PX on the viewer: to the left the
// next picture (1), to the right the one before (-1).
const SWIPE_PX = 50;
export const swipeStep = (dx: number, dy: number) => (Math.abs(dx) < SWIPE_PX || Math.abs(dy) > Math.abs(dx) ? 0 : dx < 0 ? 1 : -1);
