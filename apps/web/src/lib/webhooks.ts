// Webhooks: a secret URL that posts into a text channel (the server's
// webhooks, Discord's format). Staff manage them in WebhooksDialog; the
// URL is shown once, when it is made (only the token's hash is kept).

export type Webhook = { id: number; name: string; creator: string; created: number };

export const webhookUrl = (origin: string, id: number, token: string) => `${origin}/api/webhooks/${id}/${token}`;

export const isWebhook = (v: unknown): v is Webhook => {
  const o = v && typeof v === "object" ? (v as Record<string, unknown>) : null;
  return !!o && typeof o.id === "number" && typeof o.name === "string" && typeof o.creator === "string" && typeof o.created === "number";
};
