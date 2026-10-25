import { describe, expect, it } from "vitest";
import {
  emptyPollDraft,
  endPollRequest,
  isPoll,
  pollBody,
  pollOpen,
  pollProblem,
  share,
  showResults,
  timeLeft,
  voteRequest,
  winners,
  withPoll,
  type Poll,
} from "../src/lib/polls";

const H = 3600_000;
const poll = (over: Partial<Poll> = {}): Poll => ({
  question: "Pizza?",
  answers: [
    { text: "Yes", emoji: "🍕", votes: 2, voters: ["a", "b"] },
    { text: "No", votes: 1, voters: ["c"] },
  ],
  ends: 10 * H,
  voters: 3,
  ...over,
});

describe("isPoll", () => {
  it("takes the server's polls, not broken ones", () => {
    expect(isPoll(poll())).toBe(true);
    expect(isPoll(poll({ mine: [0], ended: 5, multiple: true, anonymous: true }))).toBe(true);
    expect(isPoll({ ...poll(), answers: [{ text: "a" }] })).toBe(false);
    expect(isPoll({ ...poll(), mine: ["0"] })).toBe(false);
    expect(isPoll({ ...poll(), ends: "soon" })).toBe(false);
  });
});

describe("the poll form", () => {
  const draft = { ...emptyPollDraft(), question: " Pizza? ", answers: [{ text: "Yes", emoji: "🍕" }, { text: " No " }, { text: "" }] };
  it("checks as the server does, empty answer fields left out", () => {
    expect(pollProblem(draft)).toBeNull();
    expect(pollProblem({ ...draft, question: " " })).toBe("Ask a question.");
    expect(pollProblem({ ...draft, question: "q".repeat(301) })).toBe("The question has at most 300 characters.");
    expect(pollProblem({ ...draft, answers: [{ text: "Yes" }, { text: "" }] })).toBe("A poll needs at least 2 answers.");
    expect(pollProblem({ ...draft, answers: [{ text: "a".repeat(56) }, { text: "b" }] })).toBe("An answer has at most 55 characters.");
  });
  it("sends the trimmed poll", () => {
    expect(pollBody(draft)).toEqual({ question: "Pizza?", answers: [{ text: "Yes", emoji: "🍕" }, { text: "No" }], multiple: false, anonymous: false, hours: 24 });
  });
  it("starts with two empty answers and a day", () => {
    expect(emptyPollDraft()).toMatchObject({ question: "", answers: [{ text: "" }, { text: "" }], hours: 24 });
  });
});

describe("showing a poll", () => {
  it("shows the counts after one's own vote or the end", () => {
    expect(showResults(poll(), 1)).toBe(false);
    expect(showResults(poll({ mine: [1] }), 1)).toBe(true);
    expect(showResults(poll({ ended: 5 }), 1)).toBe(true);
    expect(showResults(poll(), 10 * H)).toBe(true); // its time is up
    expect(pollOpen(poll(), 1)).toBe(true);
    expect(pollOpen(poll({ ended: 5 }), 1)).toBe(false);
  });
  it("gives shares of the voters and the winners", () => {
    expect(share(2, 3)).toBe(67);
    expect(share(0, 0)).toBe(0);
    expect(winners(poll())).toEqual([0]);
    expect(winners(poll({ answers: [{ text: "a", votes: 1 }, { text: "b", votes: 1 }] }))).toEqual([0, 1]);
    expect(winners(poll({ answers: [{ text: "a", votes: 0 }, { text: "b", votes: 0 }], voters: 0 }))).toEqual([]);
  });
  it("tells the time left", () => {
    expect(timeLeft(10 * H, 10 * H - 30_000)).toBe("1 minute left");
    expect(timeLeft(10 * H, 9 * H)).toBe("60 minutes left");
    expect(timeLeft(10 * H, 5 * H)).toBe("5 hours left");
    expect(timeLeft(100 * H, 0)).toBe("4 days left");
  });
});

describe("poll requests and events", () => {
  it("votes and ends", () => {
    expect(JSON.parse(voteRequest(5, [1]))).toEqual({ type: "vote", message: 5, answers: [1] });
    expect(JSON.parse(endPollRequest(5))).toEqual({ type: "end-poll", message: 5 });
  });
  it("keeps one's own votes where the server did not send them", () => {
    const m = { id: 1, poll: poll({ mine: [0] }) };
    expect(withPoll(m, poll()).poll?.mine).toEqual([0]);
    expect(withPoll(m, poll({ mine: [] })).poll?.mine).toEqual([]);
    const none: { id: number; poll?: Poll } = { id: 2 };
    expect(withPoll(none, poll()).poll?.mine).toBeUndefined();
  });
  it("never opens an ended poll again (a vote's counts that come late)", () => {
    const ended = { id: 1, poll: poll({ ended: 5 }) };
    expect(withPoll(ended, poll({ voters: 4 })).poll).toMatchObject({ ended: 5, voters: 3 });
  });
});
