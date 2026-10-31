// Webhooks: a secret URL that posts into a text channel (the server's
// webhooks, Discord's format). Staff manage them in WebhooksDialog; the
// URL is shown once, when it is made (only the token's hash is kept).

// picture: it has one (staff upload it in the dialog).
export type Webhook = { id: number; name: string; creator: string; created: number; picture?: boolean };

export const webhookUrl = (origin: string, id: number, token: string) => `${origin}/api/webhooks/${id}/${token}`;

// Its picture (also for its posts; none: 404, initials then).
export const webhookPicture = (origin: string, id: number) => `${origin}/api/webhook-picture/${id}.png`;

export const isWebhook = (v: unknown): v is Webhook => {
  const o = v && typeof v === "object" ? (v as Record<string, unknown>) : null;
  return (
    !!o &&
    typeof o.id === "number" &&
    typeof o.name === "string" &&
    typeof o.creator === "string" &&
    typeof o.created === "number" &&
    (o.picture === undefined || typeof o.picture === "boolean")
  );
};
