import { describe, expect, it } from "vitest";
import { isWebhook, webhookPicture, webhookUrl } from "../src/lib/webhooks";

describe("webhooks", () => {
  it("makes Discord's URL form on the server's address", () => {
    expect(webhookUrl("https://chat.example", 7, "tok_en-1")).toBe("https://chat.example/api/webhooks/7/tok_en-1");
  });

  it("has a picture address per webhook", () => {
    expect(webhookPicture("https://chat.example", 7)).toBe("https://chat.example/api/webhook-picture/7.png");
  });

  it("reads a listed webhook", () => {
    expect(isWebhook({ id: 1, name: "CI", creator: "mod", created: 5 })).toBe(true);
    expect(isWebhook({ id: 1, name: "CI", creator: "mod", created: 5, picture: true })).toBe(true);
    expect(isWebhook({ id: 1, name: "CI", creator: "mod", created: 5, picture: 1 })).toBe(false);
    expect(isWebhook({ id: 1, name: "CI", creator: "mod" })).toBe(false);
    expect(isWebhook({ id: "1", name: "CI", creator: "mod", created: 5 })).toBe(false);
    expect(isWebhook(null)).toBe(false);
  });
});
