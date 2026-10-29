import { describe, expect, it } from "vitest";
import { accessChannel, isLocked, memberCandidates, membersAction, privateAction, readersOf } from "../src/lib/privateChannels";
import { initialTextState, manageRequest, parseServerEvent, reduceText } from "../src/lib/textChannels";

const text = [
  { id: 1, name: "general", mumble: 10 },
  { id: 2, name: "secret", mumble: 11 },
  { id: 3, name: "old" },
];

describe("accessChannel", () => {
  it("is a voice channel itself, or a text channel's mirror", () => {
    expect(accessChannel("voice", 4, text)).toBe(4);
    expect(accessChannel("text", 2, text)).toBe(11);
    expect(accessChannel("text", 3, text)).toBeNull();
    expect(accessChannel("text", 9, text)).toBeNull();
  });
  it("is a category's own channel", () => {
    expect(accessChannel("category", 1, text, [{ id: 1, mumble: 20 }])).toBe(20);
    expect(accessChannel("category", 2, text, [{ id: 1, mumble: 20 }])).toBeNull();
  });
});

describe("private categories", () => {
  const ann = { id: 7, name: "ann" };
  const cat = { id: 7, name: "carol" };
  const state = {
    ...initialTextState,
    categories: [{ id: 1, name: "Games", mumble: 20 }],
    channels: [
      { id: 1, name: "general", mumble: 10 },
      { id: 2, name: "lobby", mumble: 11, category: 1 },
      { id: 3, name: "secret", mumble: 12, category: 1 },
    ],
    private: { 20: [ann], 12: [cat] },
    inherited: { 12: [ann] },
  };
  it("who reads a channel: its own members and the category's", () => {
    expect(readersOf(state, "text", 1)).toBeUndefined();
    expect(readersOf(state, "text", 2)).toEqual([ann]); // by the category
    expect(readersOf(state, "text", 3)).toEqual([cat, ann]);
    expect(readersOf(state, "voice", 12)).toEqual([cat, ann]);
  });
  it("locks the category and every channel in it", () => {
    expect(state.channels.map((c) => isLocked(state, c))).toEqual([false, true, true]);
  });
});

describe("memberCandidates", () => {
  it("offers registered users who are not members yet, by name (not SuperUser)", () => {
    const registered = [
      { userId: 0, name: "SuperUser" },
      { userId: 5, name: "zoe" },
      { userId: 6, name: "amy" },
      { userId: 7, name: "bob" },
    ];
    expect(memberCandidates(registered, [{ id: 7, name: "bob" }])).toEqual([
      { id: 6, name: "amy" },
      { id: 5, name: "zoe" },
    ]);
    expect(memberCandidates(null, [])).toEqual([]);
  });
});

describe("membersAction", () => {
  const bob = { id: 7, name: "bob" };
  const amy = { id: 6, name: "amy" };
  it("sends nothing without a change", () => {
    expect(membersAction(11, [bob, amy], [amy, bob])).toBeNull();
  });
  it("sends the new members, sorted", () => {
    expect(membersAction(11, [bob], [bob, amy])).toEqual({ setPrivate: 11, on: true, users: [6, 7] });
    expect(membersAction(11, [bob], [])).toEqual({ setPrivate: 11, on: true, users: [] });
  });
});

describe("privateAction", () => {
  it("turns it on or off, keeping the members", () => {
    expect(privateAction(11, false, false, [])).toBeNull();
    expect(privateAction(11, true, true, [])).toBeNull();
    expect(privateAction(11, false, true, [])).toEqual({ setPrivate: 11, on: true, users: [] });
    expect(privateAction(11, true, false, [{ id: 7, name: "bob" }])).toEqual({ setPrivate: 11, on: false, users: [] });
  });
});

describe("private channel events", () => {
  it("come with ready and on their own", () => {
    const ready = parseServerEvent(
      JSON.stringify({ type: "ready", channels: [], private: [{ channel: 11, members: [{ id: 7, name: "bob" }] }, { channel: 4 }] }),
    );
    const state = reduceText(initialTextState, ready!);
    expect(state.private).toEqual({ 11: [{ id: 7, name: "bob" }], 4: [] });
    const ev = parseServerEvent(JSON.stringify({ type: "private", channels: [] }));
    expect(reduceText(state, ev!).private).toEqual({});
    // Members by a private category, apart from the channel's own.
    const inCat = parseServerEvent(JSON.stringify({ type: "private", channels: [{ channel: 12, members: [], inherited: [{ id: 7, name: "bob" }] }] }));
    const after = reduceText(state, inCat!);
    expect([after.private, after.inherited]).toEqual([{ 12: [] }, { 12: [{ id: 7, name: "bob" }] }]);
  });
  it("drop bad entries", () => {
    expect(parseServerEvent(JSON.stringify({ type: "private", channels: [{ channel: "x" }] }))).toBeNull();
    expect(parseServerEvent(JSON.stringify({ type: "private", channels: [{ channel: 2, members: [{ id: 1 }] }] }))).toBeNull();
    expect(parseServerEvent(JSON.stringify({ type: "private", channels: [{ channel: 2, inherited: "x" }] }))).toBeNull();
    const ready = parseServerEvent(JSON.stringify({ type: "ready", channels: [], private: "x" }));
    expect(reduceText(initialTextState, ready!).private).toEqual({});
  });
  it("are asked for with set-private, create-channel and create-private-voice", () => {
    expect(JSON.parse(manageRequest({ setPrivate: 11, on: true, users: [7] }))).toEqual({ type: "set-private", voice: 11, on: true, users: [7] });
    expect(JSON.parse(manageRequest({ create: "club", private: true }))).toEqual({ type: "create-channel", name: "club", private: true });
    expect(JSON.parse(manageRequest({ createPrivateVoice: "Lounge" }))).toEqual({ type: "create-private-voice", name: "Lounge" });
    expect(JSON.parse(manageRequest({ createPrivateVoice: "Den", parent: 4 }))).toEqual({ type: "create-private-voice", name: "Den", voice: 4 });
  });
});
