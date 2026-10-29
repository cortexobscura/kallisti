import { describe, expect, it } from "vitest";
import { initialState, Perm, type Channel, type ServerState } from "@kallisti/protocol";
import { canCreateSubchannel, canManageMembers, canViewMembers, menuRights } from "../src/lib/channelManagement";
import { initialTextState, type TextState } from "../src/lib/textChannels";

const ch = (id: number, parent: number | null, position = 0): Channel => ({
  id, parent, name: `c${id}`, description: "", descriptionHash: null, position, temporary: false,
});
const server = (perms: Record<number, number>, root: number | null = null): ServerState => ({
  ...initialState,
  channels: Object.fromEntries([ch(0, null), ch(1, 0, 0), ch(2, 0, 1), ch(3, 0, 2)].map((c) => [c.id, c])),
  channelPermissions: perms,
  rootPermissions: root,
});
const text: TextState = {
  ...initialTextState,
  voiceChats: true,
  channels: [{ id: 10, name: "general" }, { id: 11, name: "random" }],
};
const W = Perm.Write;

describe("menuRights", () => {
  it("lets admins and moderators change text channels; others get a menu without changes (to mute)", () => {
    expect(menuRights(server({}, Perm.MakeChannel), text, "text", 10)).toEqual({ canEdit: true, canDelete: true, canMoveUp: false, canMoveDown: true });
    const none = { canEdit: false, canDelete: false, canMoveUp: false, canMoveDown: false };
    expect(menuRights(server({}, 0), text, "text", 10)).toEqual(none);
    expect(menuRights(server({}, Perm.MakeChannel), { ...text, voiceChats: false }, "text", 10)).toEqual(none);
    expect(menuRights(server({}, Perm.MakeChannel), text, "text", 999)).toBeNull();
  });

  it("moves a voice channel only with Write on all its siblings", () => {
    const partly = menuRights(server({ 1: W, 2: W }), text, "voice", 2);
    expect(partly).toEqual({ canEdit: true, canDelete: true, canMoveUp: false, canMoveDown: false });
    const all = menuRights(server({ 1: W, 2: W, 3: W }), text, "voice", 2);
    expect(all).toEqual({ canEdit: true, canDelete: true, canMoveUp: true, canMoveDown: true });
  });

  it("never deletes or moves root", () => {
    expect(menuRights(server({ 0: W }), text, "voice", 0)).toEqual({ canEdit: true, canDelete: false, canMoveUp: false, canMoveDown: false });
  });

  it("has nothing for a channel that is gone", () => {
    expect(menuRights(server({}), text, "voice", 99)).toBeNull();
    expect(menuRights(server({}, Perm.MakeChannel), text, "text", 99)).toBeNull();
  });
});

describe("menuRights with text channel mirrors", () => {
  it("ignores the mirrors' parent among the siblings", () => {
    const base = server({ 1: W, 2: W, 3: W });
    const mirror = (id: number, parent: number, name: string, position: number): Channel => ({
      ...ch(id, parent, position),
      name,
    });
    const s: ServerState = {
      ...base,
      channels: { ...base.channels, 5: mirror(5, 0, "Text channels", 1000), 6: mirror(6, 5, "#x", 0) },
    };
    const t: TextState = { ...text, channels: [...text.channels, { id: 12, name: "x", mumble: 6 }] };
    expect(menuRights(s, t, "voice", 3)?.canMoveDown).toBe(false);
    expect(menuRights(s, t, "voice", 3)?.canMoveUp).toBe(true);
  });
});

describe("canManageMembers", () => {
  const priv: TextState = {
    ...text,
    channels: [{ id: 10, name: "general", mumble: 20 }, { id: 11, name: "secret", mumble: 21 }],
    private: { 21: [], 2: [] },
  };
  it("is for admins and moderators, in private channels", () => {
    const staff = server({}, Perm.MakeChannel);
    expect(canManageMembers(staff, priv, "text", 11)).toBe(true);
    expect(canManageMembers(staff, priv, "voice", 2)).toBe(true);
    expect(canManageMembers(staff, priv, "text", 10)).toBe(false);
    expect(canManageMembers(staff, priv, "voice", 1)).toBe(false);
    expect(canManageMembers(server({}, 0), priv, "text", 11)).toBe(false);
  });
});

describe("canCreateSubchannel", () => {
  const M = Perm.MakeChannel;
  it("needs MakeChannel in the voice channel (staff, and channel moderators in theirs)", () => {
    expect(canCreateSubchannel(server({ 1: M }), text, 1)).toBe(true);
    expect(canCreateSubchannel(server({ 1: M }), text, 2)).toBe(false);
    expect(canCreateSubchannel(server({ 1: M }), text, 99)).toBe(false);
  });
  it("is not for root (the + makes those), temporary channels nor text channel mirrors", () => {
    expect(canCreateSubchannel(server({ 0: M }), text, 0)).toBe(false);
    const temp = server({ 1: M });
    expect(canCreateSubchannel({ ...temp, channels: { ...temp.channels, 1: { ...ch(1, 0), temporary: true } } }, text, 1)).toBe(false);
    const base = server({ 5: M, 6: M });
    const s: ServerState = { ...base, channels: { ...base.channels, 5: ch(5, 0, 1000), 6: ch(6, 5) } };
    const t: TextState = { ...text, channels: [...text.channels, { id: 12, name: "x", mumble: 6 }] };
    expect(canCreateSubchannel(s, t, 5)).toBe(false);
    expect(canCreateSubchannel(s, t, 6)).toBe(false);
  });
});

describe("canViewMembers", () => {
  const priv: TextState = {
    ...text,
    channels: [{ id: 10, name: "general", mumble: 20 }, { id: 11, name: "secret", mumble: 21 }],
    private: { 21: [{ id: 7, name: "ann" }], 2: [{ id: 8, name: "bob" }] },
  };
  const as = (userId: number, root = 0): ServerState => ({
    ...server({}, root),
    session: 1,
    users: { 1: { session: 1, name: `u${userId}`, userId, channelId: 0 } as ServerState["users"][number] },
  });
  it("is for members of a private channel who do not manage it", () => {
    expect(canViewMembers(as(7), priv, "text", 11)).toBe(true);
    expect(canViewMembers(as(8), priv, "voice", 2)).toBe(true);
    expect(canViewMembers(as(8), priv, "text", 11)).toBe(false); // not a member (the server sends no list)
    expect(canViewMembers(as(7), priv, "text", 10)).toBe(false); // not private
    expect(canViewMembers(as(7, Perm.MakeChannel), priv, "text", 11)).toBe(false); // staff: the Members dialog
    const mod = { ...priv, roles: { ...priv.roles, channelMods: { 2: [8] } } };
    expect(canViewMembers(as(8), mod, "voice", 2)).toBe(false); // its moderator: the Members dialog
    expect(canViewMembers(as(7), { ...priv, private: { 21: [] } }, "text", 11)).toBe(false); // no list sent
    // A member by the category: the channel's list too.
    expect(canViewMembers(as(9), { ...priv, private: { 21: [] }, inherited: { 21: [{ id: 9, name: "cat" }] } }, "text", 11)).toBe(true);
  });
});

describe("menuRights for a private category", () => {
  it("opens for its members and moderators too (Members…, Copy link), without changes", () => {
    const priv: TextState = {
      ...text,
      categories: [{ id: 1, name: "Games", mumble: 20 }],
      private: { 20: [{ id: 7, name: "ann" }] },
    };
    const ann: ServerState = {
      ...server({}, 0),
      session: 1,
      users: { 1: { session: 1, name: "ann", userId: 7, channelId: 0 } as ServerState["users"][number] },
    };
    const none = { canEdit: false, canDelete: false, canMoveUp: false, canMoveDown: false };
    expect(menuRights(ann, priv, "category", 1)).toEqual(none);
    expect(menuRights(server({}, 0), priv, "category", 1)).toBeNull(); // not a member: no menu
  });
});
