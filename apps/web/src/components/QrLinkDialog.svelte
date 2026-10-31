<script lang="ts">
  import { onDestroy, untrack } from "svelte";
  import { apiStatus, type QrScan, type QrState } from "../lib/accounts";

  // A scanned login code (see lib/qrLogin): the link opened, or the app's
  // scanner. Logged in, a logged-out device's code: "Log in on <device>?"
  // (the phishing warning: whoever shows it gets the account). Logged out,
  // a logged-in device's code: wait for its yes, then this device is
  // logged in (onLoggedIn). A logged-in device does not change accounts by
  // a code.
  let {
    api,
    token,
    device,
    me,
    onLoggedIn,
    onClose,
  }: {
    api: {
      qrScan: (token: string, device: string) => Promise<QrScan>;
      qrWait: (secret: string) => Promise<QrState>;
      qrAnswer: (secret: string, yes: boolean) => Promise<void>;
    };
    token: string;
    device: string; // this device, as the other one sees it
    me: string | null; // the account logged in here (null: none)
    onLoggedIn: (name: string) => void;
    onClose: () => void;
  } = $props();

  const POLL_MS = 1500;
  let dialog: HTMLDialogElement | undefined = $state();
  $effect(() => dialog?.showModal());
  let phase = $state<"scanning" | "ask" | "waiting" | "done" | "refused" | "expired" | "switch" | "error">("scanning");
  let other = $state("");
  let error = $state("");
  let timer: ReturnType<typeof setTimeout> | undefined;
  let gone = false;
  onDestroy(() => ((gone = true), clearTimeout(timer)));

  const failed = (err: unknown) => {
    const status = apiStatus(err);
    error =
      status === 404
        ? "This code is not valid any more. Make a new one."
        : status === 401
          ? "Log in on this device first, then scan the code again."
          : err instanceof Error
            ? err.message
            : String(err);
    phase = "error";
  };
  const wait = (claim: string) => {
    timer = setTimeout(async () => {
      if (gone) return;
      let st: QrState;
      try {
        st = await api.qrWait(claim);
      } catch {
        return wait(claim);
      }
      if (gone) return;
      if (st.state === "done") return void ((phase = "done"), onLoggedIn(st.name ?? ""));
      if (st.state === "refused" || st.state === "expired") return void (phase = st.state);
      wait(claim);
    }, POLL_MS);
  };
  // Once, as it opens (a code is scanned once).
  void untrack(() => api.qrScan(token, device)).then((s) => {
    if (gone) return;
    other = s.device;
    if (s.kind === "request") phase = "ask";
    else if (me) phase = "switch";
    else (phase = "waiting"), wait(s.claim);
  }, failed);
  const answer = async (yes: boolean) => {
    try {
      await api.qrAnswer(token, yes);
      phase = yes ? "done" : "refused";
    } catch (err) {
      failed(err);
    }
  };
</script>

<dialog bind:this={dialog} onclose={onClose} aria-labelledby="qr-link-title">
  <header>
    <h2 id="qr-link-title">Log in with a QR code</h2>
    <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
  </header>
  {#if phase === "scanning"}
    <p class="note">Reading the code…</p>
  {:else if phase === "ask"}
    <p><strong>Log in on {other} as {me}?</strong></p>
    <p class="warn">Only if the code is on your own screen. Whoever shows it gets your account: never scan a code that someone sent you.</p>
    <div class="buttons">
      <button type="button" class="secondary" onclick={() => answer(false)}>No</button>
      <button type="button" class="primary" onclick={() => answer(true)}>Log in</button>
    </div>
  {:else if phase === "waiting"}
    <p>Confirm on {other}…</p>
  {:else if phase === "done"}
    <p>{me ? `${other} is logged in.` : "Logged in."}</p>
  {:else if phase === "refused"}
    <p>Not logged in: the login was refused.</p>
  {:else if phase === "expired"}
    <p>The code expired. Make a new one.</p>
  {:else if phase === "switch"}
    <p>This code logs in a device without an account open. Log out here first, or scan it on that device.</p>
  {:else}
    <p role="alert">{error}</p>
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
  .warn {
    color: var(--danger);
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
