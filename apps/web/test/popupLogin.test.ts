import { describe, expect, it } from "vitest";
import { popupLogin } from "../src/lib/popupLogin";

const HOME = "https://home.example";

describe("the popup in which a Kallisti page adds this server", () => {
  it("is not one without a mark", () => {
    expect(popupLogin({ asked: null, mark: null, back: null, kept: true })).toEqual({ home: null, resumed: false, phase: null });
  });
  it("starts the login when asked", () => {
    expect(popupLogin({ asked: HOME, mark: HOME, back: null, kept: true })).toEqual({ home: HOME, resumed: false, phase: "start" });
  });
  it("fails at once when the tab keeps nothing (it could never finish)", () => {
    expect(popupLogin({ asked: HOME, mark: null, back: null, kept: false }).phase).toEqual({ error: expect.stringContaining("keeps no data") });
  });
  it("delivers, fails, or asks for the name, back from the login", () => {
    expect(popupLogin({ asked: null, mark: HOME, back: { kind: "connect" }, kept: true })).toEqual({ home: HOME, resumed: false, phase: "deliver" });
    expect(popupLogin({ asked: null, mark: HOME, back: { kind: "error", text: "no" }, kept: true }).phase).toEqual({ error: "no" });
    expect(popupLogin({ asked: null, mark: HOME, back: { kind: "register" }, kept: true })).toEqual({ home: HOME, resumed: false, phase: null });
  });
  it("comes back to the name step after a reload", () => {
    expect(popupLogin({ asked: null, mark: HOME, back: null, kept: true })).toEqual({ home: HOME, resumed: true, phase: null });
  });
  it("is no popup back from another kind of login (a normal one in this tab)", () => {
    expect(popupLogin({ asked: null, mark: HOME, back: { kind: "linked", text: "x" }, kept: true }).home).toBeNull();
    // Asked, with another kind's result in the address: no error from it.
    expect(popupLogin({ asked: HOME, mark: HOME, back: { kind: "linked", text: "x" }, kept: true }).phase).toBeNull();
  });
});
