<script lang="ts">
  import type { AccountsApi } from "../lib/accounts";

  // The popup in which a Kallisti page adds this server (see the server's
  // fediHandoff): it logs in through the user's home server (home), or
  // with their fediverse account (via: its server), then gives the token
  // to the page that opened it: only the origin the server names gets it
  // (the browser delivers nothing to another opener). No session here. A
  // login with the fediverse account asks first: any site can open this
  // popup.
  let { api, home, via = null, phase }: { api: AccountsApi; home: string; via?: string | null; phase: "start" | "deliver" | { error: string } } =
    $props();

  let error = $state<string | null>(null);
  let done = $state(false);
  // The question: this account, for this page (the server's word). Allow
  // works after a moment only (no double click lands on it, the opener
  // decides when the popup shows).
  let ask = $state<{ key: string; name: string; origin: string } | null>(null);
  let ready = $state(false);
  $effect(() => {
    if (!ask) return;
    ready = false;
    const t = setTimeout(() => (ready = true), 600);
    return () => clearTimeout(t);
  });
  const hostOf = (origin: string) => {
    try {
      return new URL(origin).host;
    } catch {
      return origin;
    }
  };
  const failed = (err: unknown) => (error = err instanceof Error ? err.message : String(err));
  const deliver = ({ token, name, origin }: { token: string; name: string; origin: string }) => {
    window.opener?.postMessage({ type: "kallisti-login", token, name }, origin);
    done = true;
    window.close();
  };
  const allow = () => {
    if (!ask || !ready) return;
    const key = ask.key;
    ask = null;
    api.fediHandoff(key).then((h) => ("token" in h ? deliver(h) : failed("The login did not finish. Please try again.")), failed);
  };
  $effect(() => {
    if (phase === "start")
      api.fediStart(via ?? home, undefined, via ? { to: home } : {}).then((url) => location.assign(url), failed);
    else if (phase === "deliver" && !window.opener) error = "The page that asked for this login is gone. Please try again there.";
    else if (phase === "deliver") api.fediHandoff().then((h) => ("token" in h ? deliver(h) : (ask = { key: h.ask, name: h.name, origin: h.origin })), failed);
    else error = phase.error;
  });
</script>

<div class="screen">
  <div class="card">
    {#if error}
      <h1>Could not log in</h1>
      <p class="error" role="alert">{error}</p>
      <button class="link" onclick={() => window.close()}>Close</button>
    {:else if ask}
      <h1>Allow {hostOf(ask.origin)}?</h1>
      <p class="sub">
        <strong>{hostOf(ask.origin)}</strong> wants to use your account <strong>{ask.name}</strong> on {location.host}. It can then read and write here
        as you.
      </p>
      <button class="primary" disabled={!ready} onclick={allow}>Allow</button>
      <button class="link" onclick={() => window.close()}>Cancel</button>
    {:else if done}
      <p class="sub">Done. You can close this window.</p>
    {:else}
      <p class="sub" aria-busy="true">{phase === "start" ? `Logging in with your account at ${hostOf(via ?? home)}…` : "Logging in…"}</p>
    {/if}
  </div>
</div>

<style>
  .screen {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    background: var(--bg-rail);
  }
  .card {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: min(400px, 90%);
    padding: 32px;
    border-radius: 8px;
    background: var(--bg-main);
  }
  h1 {
    margin: 0;
    font-size: 24px;
    text-align: center;
  }
  .sub {
    margin: 0;
    color: var(--text-muted);
    text-align: center;
  }
  .error {
    margin: 0;
    color: var(--danger);
    overflow-wrap: anywhere;
  }
  .link {
    background: none;
    color: var(--text-muted);
  }
  .primary {
    padding: 10px;
    border-radius: 4px;
    background: var(--accent-strong);
    color: white;
    font-weight: 600;
  }
</style>
