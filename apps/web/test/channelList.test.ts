import { describe, expect, it } from "vitest";
import type { ChannelNode } from "@kallisti/protocol";
import { channelListRequest, pruned, keepShown, ownView, shownBeside, voiceBeside, deviceListOf, inOwnOrder, moveOwn, noChannelList, ownChannels, ownTree, parseChannelList, parseDeviceLists, setDeviceList, toggleHidden } from "../src/lib/channelList";

const ch = (id: number) => ({ id, name: `c${id}` });
const node = (id: number, children: ChannelNode[] = []): ChannelNode => ({ channel: { id, name: `v${id}` } as ChannelNode["channel"], children });
const ids = (n: ChannelNode): unknown => [n.channel.id, n.children.map(ids)];

describe("parseChannelList", () => {
  it("reads the wire form and drops what is not a list of ids", () => {
    expect(parseChannelList({ hidden: { text: [2, "x", 3.5], voice: [4] }, order: { text: [5] } })).toEqual({
      hidden: { text: [2], voice: [4] },
      order: { text: [5], voice: [] },
    });
    expect(parseChannelList(null)).toEqual(noChannelList);
    expect(parseChannelList("nonsense")).toEqual(noChannelList);
  });
});

describe("inOwnOrder", () => {
  it("puts the ordered ones first, in the list's order; the rest after, as they were", () => {
    expect(inOwnOrder([1, 2, 3, 4, 5], (x) => x, [4, 2, 9])).toEqual([4, 2, 1, 3, 5]);
    expect(inOwnOrder([1, 2], (x) => x, [])).toEqual([1, 2]);
  });
});

describe("ownChannels", () => {
  it("leaves out hidden channels, except those kept (the open one, a mention)", () => {
    const list = [ch(1), ch(2), ch(3)];
    expect(ownChannels(list, new Set([2, 3]), [], (id) => id === 3).map((c) => c.id)).toEqual([1, 3]);
  });

  it("sorts by the own order", () => {
    expect(ownChannels([ch(1), ch(2), ch(3)], new Set(), [3, 1], () => false).map((c) => c.id)).toEqual([3, 1, 2]);
  });
});

describe("ownTree", () => {
  const tree = node(0, [node(1, [node(11), node(12)]), node(2), node(3)]);

  it("leaves out hidden voice channels with their subchannels, and sorts each level", () => {
    const list = { ...noChannelList, hidden: { text: [], voice: [1] }, order: { text: [], voice: [3] } };
    expect(ids(ownTree(tree, list, () => false))).toEqual([0, [[3, []], [2, []]]]);
  });

  it("keeps a hidden channel the user is in", () => {
    const list = { ...noChannelList, hidden: { text: [], voice: [12] } };
    expect(ids(ownTree(tree, list, (id) => id === 12))).toEqual(ids(tree));
  });
});

describe("toggleHidden", () => {
  it("hides and shows again, per kind", () => {
    const hidden = toggleHidden(noChannelList, "voice", 4);
    expect(hidden.hidden).toEqual({ text: [], voice: [4] });
    expect(toggleHidden(hidden, "voice", 4).hidden).toEqual({ text: [], voice: [] });
    expect(toggleHidden(hidden, "text", 4).hidden).toEqual({ text: [4], voice: [4] });
  });
});

describe("moveOwn", () => {
  it("moves a channel one place among its group and keeps the group's order", () => {
    const moved = moveOwn({ ...noChannelList, order: { text: [9, 2], voice: [] } }, "text", [2, 1, 3], 3, -1);
    expect(moved.order.text).toEqual([9, 2, 3, 1]);
    expect(inOwnOrder([1, 2, 3], (x) => x, moved.order.text)).toEqual([2, 3, 1]);
  });

  it("does nothing at the ends", () => {
    expect(moveOwn(noChannelList, "text", [1, 2], 1, -1)).toBe(noChannelList);
    expect(moveOwn(noChannelList, "text", [1, 2], 2, 1)).toBe(noChannelList);
  });
});

describe("the device's lists (without an account)", () => {
  it("keeps one list per server", () => {
    const one = toggleHidden(noChannelList, "text", 2);
    const all = setDeviceList({}, "https://a.example", one);
    expect(deviceListOf(all, "https://a.example")).toEqual(one);
    expect(deviceListOf(all, "https://b.example")).toEqual(noChannelList);
    expect(deviceListOf(parseDeviceLists(JSON.stringify(all)), "https://a.example")).toEqual(one);
    expect(parseDeviceLists("nonsense")).toEqual({});
  });
});

describe("channelListRequest", () => {
  it("is the server's set-channel-list", () => {
    expect(JSON.parse(channelListRequest(noChannelList))).toEqual({ type: "set-channel-list", list: noChannelList });
  });
});

describe("the channels beside one", () => {
  it("in its text group, as shown", () => {
    const groups = [{ channels: [ch(1), ch(2)] }, { channels: [ch(3), ch(4), ch(5)] }];
    expect(shownBeside(groups, 4)).toEqual([3, 4, 5]);
    expect(shownBeside(groups, 9)).toEqual([]);
  });

  it("among a voice channel's siblings", () => {
    const tree = node(0, [node(1, [node(11), node(12)]), node(2)]);
    expect(voiceBeside(tree, 12)).toEqual([11, 12]);
    expect(voiceBeside(tree, 2)).toEqual([1, 2]);
    expect(voiceBeside(tree, 0)).toEqual([]);
  });
});

describe("ownView", () => {
  it("applies the list to the text groups and the voice tree, keeping the open and mentioned ones", () => {
    const list = { hidden: { text: [1, 2], voice: [2, 3] }, order: { text: [], voice: [] } };
    const keep = { text: keepShown(1, {}), voice: keepShown(null, { 3: 1 }) };
    const view = ownView([{ channels: [ch(1), ch(2), ch(4)] }], node(0, [node(2), node(3), node(5)]), list, keep);
    expect(view.groups[0]!.channels.map((c) => c.id)).toEqual([1, 4]);
    expect(ids(view.tree!)).toEqual([0, [[3, []], [5, []]]]);
    expect(ownView([], null, list, keep).tree).toBeNull();
  });
});

describe("pruned", () => {
  it("drops ids of channels that are gone", () => {
    const list = { hidden: { text: [1, 9], voice: [4, 8] }, order: { text: [9, 1, 2], voice: [8] } };
    expect(pruned(list, new Set([1, 2]), new Set([4]))).toEqual({ hidden: { text: [1], voice: [4] }, order: { text: [1, 2], voice: [] } });
  });
});
