<script lang="ts">
  import type { Account, AccountsApi } from "../lib/accounts";
  import { isKallisti } from "../lib/desktop";
  import { serverOrigin } from "../lib/serverList";

  // Another Kallisti server for the rail (several servers in one client):
  // its address, then a login there. New accounts and guests are made on
  // that server's own site (it refuses other sites). The login's token
  // stays here until onAdded takes it (a cancelled or late login is logged
  // out again).
  let {
    home,
    taken,
    apiFor,
    onAdded,
    onClose,
  }: {
    home: string;
    // Origins already in the rail.
    taken: readonly string[];
    // A server's account API whose token stays in memory (see token()).
    apiFor: (origin: string) => { api: AccountsApi; token: () => string | null };
    onAdded: (origin: string, account: Account, token: string) => void;
    onClose: () => void;
  } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  let address = $state("");
  let found = $state<{ origin: string } & ReturnType<typeof apiFor> | null>(null);
  let name = $state("");
  let password = $state("");
  let error = $state<string | null>(null);
  let busy = $state(false);
  // Each check and login gets a number: Back, Close and a new try make
  // older ones count for nothing.
  let attempt = 0;
  $effect(() => dialog?.showModal());
  $effect(() => () => void attempt++);

  const back = () => {
    attempt++;
    found = null;
    name = password = "";
    error = null;
    busy = false;
  };

  const check = async (event: SubmitEvent) => {
    event.preventDefault();
    error = null;
    const parsed = serverOrigin(address, home);
    if ("error" in parsed) return void (error = parsed.error);
    const { origin } = parsed;
    if (taken.includes(origin)) return void (error = "This server is already in the list.");
    const mine = ++attempt;
    busy = true;
    const kind = await isKallisti(origin);
    if (mine !== attempt) return;
    busy = false;
    if (kind === "unreachable") error = `Cannot reach ${origin}. Check the address, or try again.`;
    else if (kind === "other") error = `${origin} is not a Kallisti server.`;
    else found = { origin, ...apiFor(origin) };
  };

  const logIn = async (event: SubmitEvent) => {
    event.preventDefault();
    const target = found;
    if (!target) return;
    const mine = ++attempt;
    error = null;
    busy = true;
    const cancelled = () => mine !== attempt;
    try {
      const result = await target.api.login(name.trim(), password);
      // A Mumble registration has no session for other sites (the server
      // refuses it there; an older one may not).
      if (result.mumble) throw new Error("Log in with this name on the server's own site.");
      const account = await target.api.me();
      const token = target.token();
      if (cancelled()) {
        if (token) void target.api.logout();
        return;
      }
      if (!account || !token) throw new Error("The login did not last. Please try again.");
      onAdded(target.origin, account, token);
      dialog?.close();
    } catch (err) {
      if (!cancelled()) error = err instanceof Error ? err.message : String(err);
    } finally {
      if (!cancelled()) busy = false;
    }
  };
</script>

<dialog bind:this={dialog} onclose={() => (attempt++, onClose())} aria-labelledby="add-server-title">
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
        <button type="button" class="secondary" onclick={back}>Back</button>
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
