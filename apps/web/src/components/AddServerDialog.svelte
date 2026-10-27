<script lang="ts">
  import type { Account, AccountsApi } from "../lib/accounts";
  import { isKallisti, kallistiOrigin } from "../lib/desktop";

  // Another Kallisti server for the rail (several servers in one client):
  // its address, then a login there. New accounts and guests are made on
  // that server's own site (it refuses other sites).
  let {
    taken,
    apiFor,
    onAdded,
    onClose,
  }: {
    // Origins already in the rail (the home server too).
    taken: readonly string[];
    // The account API of a server, with its session token kept on this device.
    apiFor: (origin: string) => AccountsApi;
    onAdded: (origin: string, api: AccountsApi, account: Account) => void;
    onClose: () => void;
  } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  let address = $state("");
  let found = $state<{ origin: string; api: AccountsApi } | null>(null);
  let name = $state("");
  let password = $state("");
  let error = $state<string | null>(null);
  let busy = $state(false);
  $effect(() => dialog?.showModal());

  const check = async (event: SubmitEvent) => {
    event.preventDefault();
    error = null;
    const origin = kallistiOrigin(address);
    if (!origin) return void (error = "Enter a server address, for example chat.example.org");
    if (taken.includes(origin)) return void (error = "This server is already in the list.");
    busy = true;
    const kind = await isKallisti(origin);
    busy = false;
    if (kind === "unreachable") error = `Cannot reach ${origin}. Check the address, or try again.`;
    else if (kind === "other") error = `${origin} is not a Kallisti server.`;
    else found = { origin, api: apiFor(origin) };
  };

  const logIn = async (event: SubmitEvent) => {
    event.preventDefault();
    if (!found) return;
    error = null;
    busy = true;
    try {
      const result = await found.api.login(name.trim(), password);
      // A Mumble registration has no session for other sites (the server
      // refuses it there; an older one may not).
      if (result.mumble) throw new Error("Log in with this name on the server's own site.");
      const account = await found.api.me();
      if (!account) throw new Error("The login did not last. Please try again.");
      onAdded(found.origin, found.api, account);
      dialog?.close();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
    }
  };
</script>

<dialog bind:this={dialog} onclose={onClose} aria-labelledby="add-server-title">
  <header>
    <h2 id="add-server-title">Add a server</h2>
    <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
  </header>
  {#if !found}
    <form onsubmit={check}>
      <label>
        Server address
        <!-- svelte-ignore a11y_autofocus -->
        <input bind:value={address} required placeholder="chat.example.org" autocomplete="url" autofocus />
      </label>
      <p class="hint">A Kallisti server where you have an account.</p>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <footer>
        <button type="button" class="secondary" onclick={() => dialog?.close()}>Cancel</button>
        <button type="submit" class="primary" disabled={busy || !address.trim()}>{busy ? "Looking…" : "Continue"}</button>
      </footer>
    </form>
  {:else}
    <form onsubmit={logIn}>
      <p class="server">{found.origin}</p>
      <label>
        Name
        <!-- svelte-ignore a11y_autofocus -->
        <input bind:value={name} required maxlength="32" autocomplete="username" autofocus />
      </label>
      <label>
        Password
        <input bind:value={password} type="password" required autocomplete="current-password" />
      </label>
      <p class="hint">No account there yet? Make one on <a href={found.origin} target="_blank" rel="noopener noreferrer">its own site</a>.</p>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <footer>
        <button type="button" class="secondary" onclick={() => ((found = null), (error = null))}>Back</button>
        <button type="submit" class="primary" disabled={busy || !name.trim() || !password}>{busy ? "Logging in…" : "Log in"}</button>
      </footer>
    </form>
  {/if}
</dialog>

<style>
  dialog {
    width: min(440px, 92vw);
    padding: 24px;
    border: none;
    border-radius: 8px;
    background: var(--bg-main);
    color: var(--text);
  }
  dialog::backdrop {
    background: rgb(0 0 0 / 0.6);
  }
  header,
  footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  header {
    padding-bottom: 16px;
  }
  footer {
    justify-content: flex-end;
    gap: 8px;
    margin-top: 8px;
  }
  form {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  h2 {
    margin: 0;
    font-size: 18px;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-muted);
  }
  input {
    padding: 10px;
    border: none;
    border-radius: 4px;
    background: var(--bg-input);
    color: var(--text);
    font-size: 15px;
    text-transform: none;
  }
  .hint,
  .server {
    margin: 0;
    color: var(--text-muted);
    font-size: 13px;
  }
  .server {
    color: var(--text);
    font-weight: 600;
  }
  .error {
    margin: 0;
    color: var(--danger);
    font-size: 14px;
  }
  .close {
    font-size: 24px;
    line-height: 1;
  }
  .primary,
  .secondary {
    padding: 8px 16px;
    border-radius: 4px;
  }
  .primary {
    background: var(--accent-strong);
    color: #fff;
  }
  .primary:disabled {
    opacity: 0.5;
  }
  .secondary {
    background: var(--bg-input);
  }
</style>
