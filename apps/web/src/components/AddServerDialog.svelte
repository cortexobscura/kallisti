<script lang="ts">
  import { untrack } from "svelte";
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
    homeCookie = true,
    taken,
    server,
    address: startAddress = "",
    apiFor,
    onAdded,
    onClose,
  }: {
    home: string;
    // The home login is a cookie (the web app): see serverOrigin.
    homeCookie?: boolean;
    // Origins already in the rail.
    taken: readonly string[];
    // A server of the list without a login on this device: straight to its
    // login, with the name it had.
    server?: { origin: string; name: string };
    // An address to check at once (an invite link's server).
    address?: string;
    // A server's account API whose token stays in memory (see token()),
    // starting with this one.
    apiFor: (origin: string, token?: string) => { api: AccountsApi; token: () => string | null };
    // False: not taken (no longer logged in at home, or added meanwhile).
    onAdded: (origin: string, account: Account, token: string) => boolean;
    onClose: () => void;
  } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  // svelte-ignore state_referenced_locally
  const preset = server;
  // svelte-ignore state_referenced_locally
  let address = $state(startAddress);
  let found = $state<{ origin: string } | null>(preset ? { origin: preset.origin } : null);
  let name = $state(preset?.name ?? "");
  // The name and password fields: hidden until asked for (no home account
  // to use in the desktop app: shown).
  // svelte-ignore state_referenced_locally
  let withPassword = $state(!homeCookie);
  let password = $state("");
  let error = $state<string | null>(null);
  let busy = $state(false);
  // Each check and login gets a number: Back, Close and a new try make
  // older ones count for nothing, and log out a login they got (each login
  // has its own token: a late one cannot touch a newer one).
  let attempt = 0;
  let holding: { logout: () => void } | null = null;
  const cancel = () => {
    attempt++;
    holding?.logout();
    holding = null;
  };
  $effect(() => dialog?.showModal());
  // Once, after the setup (check is below).
  $effect(() => {
    if (startAddress && !preset) untrack(() => void check());
  });
  $effect(() => cancel);

  const back = () => {
    cancel();
    found = null;
    name = password = "";
    withPassword = !homeCookie;
    error = null;
    busy = false;
  };

  const check = async (event?: SubmitEvent) => {
    event?.preventDefault();
    error = null;
    const parsed = serverOrigin(address, home, homeCookie);
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
    else found = { origin };
  };

  // With the home account, no typing (web only): the server's page in a
  // popup logs in through the home server and posts the token back (see
  // Connect there). Only that popup, from that server's origin, counts.
  const viaHome = () => {
    if (!found) return;
    const origin = found.origin;
    cancel();
    const mine = attempt;
    error = null;
    const popup = window.open(`${origin}/?kallisti-connect=${encodeURIComponent(home)}`, "kallisti-connect", "popup,width=480,height=720");
    if (!popup) return void (error = "The browser blocked the login window. Allow pop-ups for this site, or log in with a password.");
    busy = true;
    let result = false;
    const stop = () => {
      window.removeEventListener("message", onMessage);
      clearInterval(watch);
      if (mine === attempt) busy = false;
    };
    const onMessage = (e: MessageEvent) => {
      const d = e.data as { type?: unknown; token?: unknown; name?: unknown } | null;
      if (e.origin !== origin || e.source !== popup || d?.type !== "kallisti-login" || typeof d.token !== "string" || typeof d.name !== "string") return;
      result = true;
      stop();
      if (!onAdded(origin, { name: d.name, temporary: false }, d.token)) {
        void apiFor(origin, d.token).api.logout();
        error = "This server could not be added (it may be in the list already, or you are no longer logged in).";
        return;
      }
      dialog?.close();
    };
    const watch = setInterval(() => {
      if (!popup.closed) return;
      stop();
      if (!result && mine === attempt) error = "The login window was closed before it finished.";
    }, 500);
    window.addEventListener("message", onMessage);
    holding = { logout: () => (stop(), popup.close()) };
  };

  const logIn = async (event: SubmitEvent) => {
    event.preventDefault();
    if (!found) return;
    const origin = found.origin;
    cancel();
    const mine = attempt;
    const { api, token } = apiFor(origin);
    // Until it is taken, this login's token is ours to log out.
    const mineHolds = { logout: () => void (token() && api.logout()) };
    holding = mineHolds;
    const drop = () => {
      if (holding === mineHolds) holding = null;
      mineHolds.logout();
    };
    const cancelled = () => mine !== attempt;
    error = null;
    busy = true;
    try {
      const result = await api.login(name.trim(), password);
      if (cancelled()) return drop();
      // A Mumble registration has no session for other sites (the server
      // refuses it there; an older one may not).
      if (result.mumble) throw new Error("Log in with this name on the server's own site.");
      const account = await api.me();
      if (cancelled()) return drop();
      const t = token();
      if (!account || !t) throw new Error("The login did not last. Please try again.");
      holding = null;
      if (!onAdded(origin, account, t)) {
        mineHolds.logout();
        throw new Error("This server could not be added (it may be in the list already, or you are no longer logged in).");
      }
      dialog?.close();
    } catch (err) {
      if (holding === mineHolds) drop();
      if (!cancelled()) error = err instanceof Error ? err.message : String(err);
    } finally {
      if (!cancelled()) busy = false;
    }
  };
</script>

<dialog bind:this={dialog} onclose={() => (cancel(), onClose())} aria-labelledby="add-server-title">
  <header>
    <h2 id="add-server-title">{preset ? `Log in to ${new URL(preset.origin).host}` : "Add a server"}</h2>
    <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
  </header>
  {#if !found}
    <form onsubmit={check}>
      <label>
        Server address
        <!-- svelte-ignore a11y_autofocus -->
        <input bind:value={address} required placeholder="chat.example.org" autocomplete="url" autofocus />
      </label>
      <p class="hint">Any Kallisti server: with your account here, or one you have there.</p>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <footer>
        <button type="button" class="secondary" onclick={() => dialog?.close()}>Cancel</button>
        <button type="submit" class="primary" disabled={busy || !address.trim()}>{busy ? "Looking…" : "Continue"}</button>
      </footer>
    </form>
  {:else}
    <form onsubmit={logIn}>
      <p class="server">{found.origin}</p>
      <!-- The home account first (an account there is made if needed); a
           password there only when asked for (and in the desktop app). -->
      {#if homeCookie}
        <p class="hint">You join with your account at {new URL(home).host}. New there? Your account is made at once.</p>
        <button type="button" class="primary" disabled={busy} onclick={viaHome}>Continue with your account at {new URL(home).host}</button>
      {/if}
      {#if withPassword}
        <label>
          Name
          <!-- svelte-ignore a11y_autofocus -->
          <input bind:value={name} required maxlength="128" autocomplete="username" autofocus />
        </label>
        <label>
          Password
          <input bind:value={password} type="password" required autocomplete="current-password" />
        </label>
        <p class="hint">No account there yet? Make one on <a href={found.origin} target="_blank" rel="noopener noreferrer">its own site</a>.</p>
      {:else}
        <button type="button" class="link" onclick={() => (withPassword = true)}>I have a password for an account there</button>
      {/if}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <footer>
        {#if preset}
          <button type="button" class="secondary" onclick={() => dialog?.close()}>Cancel</button>
        {:else}
          <button type="button" class="secondary" onclick={back}>Back</button>
        {/if}
        {#if withPassword}
          <button type="submit" class="primary" disabled={busy || !name.trim() || !password}>{busy ? "Logging in…" : "Log in"}</button>
        {/if}
      </footer>
    </form>
  {/if}
</dialog>

<style>
  .link {
    align-self: center;
    background: none;
    color: var(--text-muted);
    text-decoration: underline;
  }
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
