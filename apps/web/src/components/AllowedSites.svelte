<script lang="ts">
  import type { AccountActions, AllowedSite } from "../lib/accounts";

  // The settings' list of other servers this account logs in to (Kallisti
  // as a login provider); one taken back asks again at its next login.
  // Nothing shown while there are none.
  let { sites }: { sites: NonNullable<AccountActions["sites"]> } = $props();

  let list = $state<AllowedSite[]>([]);
  let error = $state<string | null>(null);
  let busy = $state<string | null>(null);
  $effect(() => {
    void sites.list().then(
      (got) => (list = got),
      () => {},
    );
  });
  const revoke = async (redirect: string) => {
    busy = redirect;
    error = null;
    try {
      await sites.revoke(redirect);
      list = list.filter((s) => s.redirect !== redirect);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = null;
    }
  };
</script>

{#if list.length > 0}
  <div class="account">
    <span class="label">Servers you log in to with this account</span>
    <ul>
      {#each list as s (s.redirect)}
        <li>
          <span class="site">{new URL(s.site).host}</span>
          <button type="button" class="secondary" disabled={busy !== null} onclick={() => revoke(s.redirect)}>
            {busy === s.redirect ? "Removing…" : "Remove"}
          </button>
        </li>
      {/each}
    </ul>
    {#if error}<p class="hint" role="alert">{error}</p>{/if}
    <p class="hint">A removed server asks for your OK again at your next login there. Your account there stays.</p>
  </div>
{/if}

<style>
  .account {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-top: 12px;
    margin-bottom: 16px;
    border-top: 1px solid var(--border-subtle);
  }
  .label {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-muted);
  }
  ul {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .site {
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  .secondary {
    padding: 6px 12px;
    border-radius: 6px;
    background: var(--bg-input);
    font-weight: 600;
  }
  .hint {
    margin: 0;
    font-size: 12px;
    color: var(--text-faint);
  }
</style>
