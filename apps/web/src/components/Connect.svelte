<script lang="ts">
  import type { AccountsApi } from "../lib/accounts";

  // The popup in which a Kallisti page adds this server (see the server's
  // fediHandoff): it logs in through the user's home server (home), then
  // gives the token to the page that opened it: only the home server's
  // origin, as the server says, gets it (the browser delivers nothing to
  // another opener). No session here.
  let { api, home, phase }: { api: AccountsApi; home: string; phase: "start" | "deliver" | { error: string } } = $props();

  let error = $state<string | null>(null);
  let done = $state(false);
  const homeHost = (() => {
    try {
      return new URL(home).host;
    } catch {
      return home;
    }
  })();
  $effect(() => {
    if (phase === "start")
      api.fediStart(home, undefined, true).then(
        (url) => location.assign(url),
        (err) => (error = err instanceof Error ? err.message : String(err)),
      );
    else if (phase === "deliver" && !window.opener) error = "The page that asked for this login is gone. Please try again there.";
    else if (phase === "deliver")
      api.fediHandoff().then(
        ({ token, name, origin }) => {
          window.opener?.postMessage({ type: "kallisti-login", token, name }, origin);
          done = true;
          window.close();
        },
        (err) => (error = err instanceof Error ? err.message : String(err)),
      );
    else error = phase.error;
  });
</script>

<div class="screen">
  <div class="card">
    {#if error}
      <h1>Could not log in</h1>
      <p class="error" role="alert">{error}</p>
      <button class="link" onclick={() => window.close()}>Close</button>
    {:else if done}
      <p class="sub">Done. You can close this window.</p>
    {:else}
      <p class="sub" aria-busy="true">Logging in with your account at {homeHost}…</p>
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
</style>
