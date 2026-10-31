<script lang="ts">
  import { onDestroy } from "svelte";
  import type { QrStart, QrState } from "../lib/accounts";
  import { qrLink } from "../lib/qrLogin";
  import QrCode from "./QrCode.svelte";

  // Shows a login code (see lib/qrLogin) and follows it. Logged in (an
  // offer): after a scan, asks "Log in on <device>?". Logged out (a
  // request): waits for the phone's yes, then this device is logged in
  // (onLoggedIn).
  let {
    api,
    origin,
    device,
    onLoggedIn,
    onClose,
  }: {
    api: {
      qrStart: (device: string) => Promise<QrStart>;
      qrWait: (secret: string) => Promise<QrState>;
      qrAnswer: (secret: string, yes: boolean) => Promise<void>;
    };
    origin: string; // the server's address, in the code
    device: string; // this device, as the other one sees it
    onLoggedIn?: (name: string) => void;
    onClose: () => void;
  } = $props();

  const POLL_MS = 1500;
  let dialog: HTMLDialogElement | undefined = $state();
  $effect(() => dialog?.showModal());
  let code = $state<QrStart | null>(null);
  let phase = $state<"loading" | "showing" | "ask" | "confirm" | "approving" | "done" | "refused" | "expired" | "error">("loading");
  let other = $state(""); // the other device
  let error = $state("");
  let timer: ReturnType<typeof setTimeout> | undefined;
  let gone = false;
  onDestroy(() => ((gone = true), clearTimeout(timer)));

  const follow = (c: QrStart) => {
    timer = setTimeout(async () => {
      if (gone || code !== c) return;
      let st: QrState;
      try {
        st = await api.qrWait(c.owner);
      } catch {
        return follow(c); // a lost request: the next one
      }
      if (gone || code !== c) return;
      if (st.device) other = st.device;
      switch (st.state) {
        case "scanned":
          if (phase === "showing") phase = c.kind === "offer" ? "ask" : "confirm";
          break;
        case "approved":
          if (c.kind === "offer") phase = "approving";
          break;
        case "refused":
          return void (phase = "refused");
        case "done": // a request: this device is logged in
          phase = "done";
          return onLoggedIn?.(st.name ?? "");
        case "expired":
          // An offer the phone took: done.
          return void (phase = phase === "approving" ? "done" : "expired");
      }
      follow(c);
    }, POLL_MS);
  };
  const start = async () => {
    clearTimeout(timer);
    phase = "loading";
    other = "";
    try {
      code = await api.qrStart(device);
      phase = "showing";
      follow(code);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      phase = "error";
    }
  };
  void start();
  const answer = async (yes: boolean) => {
    if (!code) return;
    try {
      await api.qrAnswer(code.owner, yes);
      phase = yes ? "approving" : "refused";
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      phase = "error";
    }
  };
  const title = $derived(code?.kind === "request" ? "Log in with a QR code" : "Log in on another device");
</script>

<dialog bind:this={dialog} onclose={onClose} aria-labelledby="qr-show-title">
  <header>
    <h2 id="qr-show-title">{title}</h2>
    <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
  </header>
  {#if phase === "loading"}
    <p class="note">Making a code…</p>
  {:else if phase === "showing" && code}
    <div class="code"><QrCode text={qrLink(origin, code.token)} label="Login code" /></div>
    <p class="note">
      {#if code.kind === "offer"}
        Scan this code with the Kallisti app (or the camera) on the device you want to log in. You confirm here.
      {:else}
        Scan this code with the Kallisti app (or the camera) on a phone where you are logged in, and confirm there.
      {/if}
      The code is valid for 2 minutes.
    </p>
  {:else if phase === "ask"}
    <p><strong>Log in on {other}?</strong></p>
    <p class="note">Only if you scanned the code yourself, just now.</p>
    <div class="buttons">
      <button type="button" class="secondary" onclick={() => answer(false)}>No</button>
      <button type="button" class="primary" onclick={() => answer(true)}>Log in</button>
    </div>
  {:else if phase === "confirm"}
    <p>Confirm on {other}…</p>
  {:else if phase === "approving"}
    <p>Logging in on {other}…</p>
  {:else if phase === "done"}
    <p>{code?.kind === "offer" ? `Logged in on ${other}.` : "Logged in."}</p>
  {:else if phase === "refused"}
    <p>Not logged in: the login was refused.</p>
    <div class="buttons"><button type="button" class="primary" onclick={start}>New code</button></div>
  {:else if phase === "expired"}
    <p>The code expired.</p>
    <div class="buttons"><button type="button" class="primary" onclick={start}>New code</button></div>
  {:else}
    <p role="alert">{error}</p>
    <div class="buttons"><button type="button" class="primary" onclick={start}>Try again</button></div>
  {/if}
</dialog>

<style>
  dialog {
    width: min(360px, 92vw);
    padding: 24px;
    border: none;
    border-radius: 8px;
    background: var(--bg-main);
    color: var(--text);
  }
  dialog::backdrop {
    background: rgb(0 0 0 / 0.6);
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 8px;
  }
  h2 {
    margin: 0;
    font-size: 18px;
  }
  .close {
    font-size: 24px;
    line-height: 1;
    color: var(--text-muted);
  }
  p {
    margin: 0 0 12px;
    font-size: 14px;
  }
  .note {
    color: var(--text-muted);
  }
  .code {
    display: grid;
    place-items: center;
    margin: 8px 0 16px;
  }
  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
  .buttons button {
    padding: 10px 16px;
    border-radius: 4px;
    font-weight: 600;
  }
  .primary {
    background: var(--accent-strong);
    color: white;
  }
  .secondary {
    background: none;
    color: var(--text);
  }
</style>
