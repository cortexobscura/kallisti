import { describe, expect, it } from "vitest";
import { connectUrl, connectWays, popupLogin } from "../src/lib/popupLogin";

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

describe("connectWays", () => {
  const fedi = { handle: "@Lars@shitposter.world", public: false, host: "shitposter.world" };
  it("a local account goes through its home server", () => {
    expect(connectWays({ name: "lars" })).toEqual({ home: true, fedi: null });
  });
  it("a linked local account may also use its fediverse account", () => {
    expect(connectWays({ name: "lars", fedi })).toEqual({ home: true, fedi: { handle: "@Lars@shitposter.world", host: "shitposter.world" } });
  });
  it("an account named after its fediverse home only with that account", () => {
    expect(connectWays({ name: "Lars@shitposter.world", fedi })).toEqual({ home: false, fedi: { handle: "@Lars@shitposter.world", host: "shitposter.world" } });
  });
  it("no host from an older server: no fediverse way", () => {
    expect(connectWays({ name: "lars", fedi: { handle: "@lars@x.example", public: false } })).toEqual({ home: true, fedi: null });
  });
});

describe("connectUrl", () => {
  it("names the home server, and the fediverse server to log in with", () => {
    expect(connectUrl("https://b.example", "https://a.example")).toBe("https://b.example/?kallisti-connect=https%3A%2F%2Fa.example");
    expect(connectUrl("https://b.example", "https://a.example", "shitposter.world")).toBe(
      "https://b.example/?kallisti-connect=https%3A%2F%2Fa.example&kallisti-via=shitposter.world",
    );
  });
});
