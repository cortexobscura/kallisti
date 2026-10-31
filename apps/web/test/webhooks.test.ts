import { describe, expect, it } from "vitest";
import { isWebhook, webhookUrl } from "../src/lib/webhooks";

describe("webhooks", () => {
  it("makes Discord's URL form on the server's address", () => {
    expect(webhookUrl("https://chat.example", 7, "tok_en-1")).toBe("https://chat.example/api/webhooks/7/tok_en-1");
  });

  it("reads a listed webhook", () => {
    expect(isWebhook({ id: 1, name: "CI", creator: "mod", created: 5 })).toBe(true);
    expect(isWebhook({ id: 1, name: "CI", creator: "mod" })).toBe(false);
    expect(isWebhook({ id: "1", name: "CI", creator: "mod", created: 5 })).toBe(false);
    expect(isWebhook(null)).toBe(false);
  });
});
