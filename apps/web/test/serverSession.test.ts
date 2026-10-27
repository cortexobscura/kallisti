// @vitest-environment jsdom
import { flushSync } from "svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createServerSession } from "../src/lib/serverSession.svelte";
import { fakeConnection, type FakeConnection } from "./helpers/fakeConnection.svelte";

const credentials = { username: "alice" };

const setup = (local = false) => {
  const made: FakeConnection[] = [];
  const session = createServerSession<FakeConnection>({
    site: () => "https://chat.example",
    urlFor: async () => ({ url: "wss://chat.example/ws", local }),
    connect: (target) => {
      const c = fakeConnection(target.url);
      made.push(c);
      return c;
    },
    tabStorage: () => undefined,
  });
  return { session, made };
};
// Lets awaited promises and effects run.
const settle = async () => {
  await vi.advanceTimersByTimeAsync(0);
  flushSync();
};

describe("a server session", () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it("connects, and a later start wins over a slower earlier one", async () => {
    const { session, made } = setup();
    void session.start(credentials);
    void session.start({ username: "bob" });
    await settle();
    expect(made).toHaveLength(1);
    expect(session.connection).toBe(made[0]);
    expect(session.credentials?.username).toBe("bob");
    session.dispose();
  });

  it("comes back quietly after a loss: an attempt each second, then the new connection", async () => {
    const { session, made } = setup();
    await session.start(credentials);
    await settle();
    made[0]!.sync();
    made[0]!.drop();
    flushSync();
    expect(session.lost).toBe(true);
    expect(session.notice).toBe("none");
    await vi.advanceTimersByTimeAsync(1000);
    flushSync();
    expect(made).toHaveLength(2);
    made[1]!.sync();
    flushSync();
    expect(session.connection).toBe(made[1]);
    expect(session.lost).toBe(false);
    session.dispose();
  });

  it("shows a small sign after 2 s and the banner after 10 s", async () => {
    const { session, made } = setup();
    await session.start(credentials);
    await settle();
    made[0]!.sync();
    made[0]!.drop();
    flushSync();
    await vi.advanceTimersByTimeAsync(2500);
    flushSync();
    expect(session.notice).toBe("sign");
    await vi.advanceTimersByTimeAsync(9000);
    flushSync();
    expect(session.notice).toBe("banner");
    session.dispose();
  });

  it("stops at a refusal by the server's rules and shows it", async () => {
    const { session, made } = setup();
    await session.start(credentials);
    await settle();
    made[0]!.sync();
    made[0]!.drop();
    flushSync();
    await vi.advanceTimersByTimeAsync(1000);
    flushSync();
    made[1]!.drop("join as a guest on the server's own site");
    flushSync();
    expect(session.connection).toBe(made[1]);
    expect(session.error).toBe("join as a guest on the server's own site");
    await vi.advanceTimersByTimeAsync(60_000);
    expect(made).toHaveLength(2);
    session.dispose();
  });

  it("shows the desktop bridge's reason only for its own connections", async () => {
    for (const local of [true, false]) {
      const { session, made } = setup(local);
      await session.start(credentials);
      await settle();
      made[0]!.drop("cannot reach voice.example:64738: refused");
      flushSync();
      expect(session.error).toBe(local ? "cannot reach voice.example:64738: refused" : "Lost the connection to the server.");
      session.dispose();
    }
  });

  it("leaves: no connection and no more attempts", async () => {
    const { session, made } = setup();
    await session.start(credentials);
    await settle();
    made[0]!.sync();
    made[0]!.drop();
    flushSync();
    session.leave();
    await vi.advanceTimersByTimeAsync(60_000);
    expect(session.connection).toBeNull();
    expect(made).toHaveLength(1);
    session.dispose();
  });
});
