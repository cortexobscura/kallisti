<script lang="ts">
  import type { AccountsApi } from "../lib/accounts";

  // Another server asks who you are here (Kallisti as a login provider):
  // allowed once per site, later logins there go through at once.
  let { api, query, serverName }: { api: AccountsApi; query: string; serverName: string } = $props();

  type Ask = Awaited<ReturnType<AccountsApi["oauthRequest"]>>;
  let ask = $state<Ask | null>(null);
  let error = $state<string | null>(null);
  let busy = $state(false);
  $effect(() => {
    api.oauthRequest(query).then(
      (a) => (ask = a),
      (err) => (error = err instanceof Error ? err.message : String(err)),
    );
  });
  const host = (site: string) => new URL(site).host;
  const allow = async () => {
    busy = true;
    try {
      location.assign(await api.oauthApprove(query));
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      busy = false;
    }
  };
</script>

<div class="screen">
  <div class="card">
    {#if ask}
      <h1>Log in to {host(ask.site)}</h1>
      <p class="sub">
        <strong>{host(ask.site)}</strong> wants to know who you are on {serverName}. It sees your name,
        <strong>{ask.account}</strong>, and nothing else.
      </p>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <button class="primary" disabled={busy} onclick={allow}>Allow</button>
      <button class="link" disabled={busy} onclick={() => location.assign(ask!.deny)}>Cancel</button>
    {:else if error}
      <h1>Cannot log in there</h1>
      <p class="error" role="alert">{error}</p>
      <button class="link" onclick={() => location.assign("/")}>Back to {serverName}</button>
    {:else}
      <p class="sub" aria-busy="true">Loading…</p>
    {/if}
  </div>
</div>

<style>
  .screen {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    padding: var(--safe-top) var(--safe-right) var(--safe-bottom) var(--safe-left);
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
    overflow-wrap: anywhere;
  }
  .sub {
    margin: 0;
    color: var(--text-muted);
    text-align: center;
    overflow-wrap: anywhere;
  }
  .error {
    margin: 0;
    color: var(--danger);
  }
  .primary {
    padding: 10px;
    border-radius: 4px;
    background: var(--accent-strong);
    color: white;
    font-weight: 600;
  }
  .link {
    background: none;
    color: var(--text-muted);
  }
</style>
