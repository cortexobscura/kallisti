import { describe, expect, it } from "vitest";
import { gridCells, splitFiles, step, swipeStep } from "../src/lib/fileGrid";

const rows = (n: number) => gridCells(n).map((c) => `${c.cols}x${c.rows}`).join(" ");

describe("gridCells", () => {
  it("lays pictures out as in Discord (6 columns)", () => {
    expect(rows(2)).toBe("3x1 3x1");
    expect(rows(3)).toBe("4x2 2x1 2x1"); // one big, two small
    expect(rows(4)).toBe("3x1 3x1 3x1 3x1");
    expect(rows(5)).toBe("3x1 3x1 2x1 2x1 2x1");
    expect(rows(6)).toBe("2x1 2x1 2x1 2x1 2x1 2x1");
    expect(rows(7)).toBe("6x1 2x1 2x1 2x1 2x1 2x1 2x1");
    expect(rows(10)).toBe("6x1" + " 2x1".repeat(9));
  });
  it("gives one cell per picture", () => {
    for (let n = 2; n <= 10; n++) expect(gridCells(n)).toHaveLength(n);
  });
});

describe("splitFiles", () => {
  it("puts two or more pictures in the grid, the rest below in order", () => {
    const f = (name: string, view: string) => ({ name, view });
    const viewOf = (u: { view: string }) => u.view;
    const split = splitFiles([f("a", "image"), f("b", "file"), f("c", "image"), f("d", "video")], viewOf);
    expect(split.grid.map((u) => u.name)).toEqual(["a", "c"]);
    expect(split.rest.map((u) => u.name)).toEqual(["b", "d"]);
    // One picture is shown as before, not in a grid.
    expect(splitFiles([f("a", "image"), f("b", "file")], viewOf)).toEqual({ grid: [], rest: [f("a", "image"), f("b", "file")] });
  });
});

describe("viewer steps", () => {
  it("stays within the pictures", () => {
    expect(step(0, -1, 3)).toBe(0);
    expect(step(0, 1, 3)).toBe(1);
    expect(step(2, 1, 3)).toBe(2);
  });
  it("takes a sideways swipe, not a short or upright one", () => {
    expect(swipeStep(-80, 10)).toBe(1); // to the left: the next one
    expect(swipeStep(80, -5)).toBe(-1);
    expect(swipeStep(20, 0)).toBe(0);
    expect(swipeStep(70, 90)).toBe(0);
  });
});
