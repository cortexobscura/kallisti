<script lang="ts">
  import { onDestroy, untrack } from "svelte";
  import { apiStatus, type QrScan, type QrState } from "../lib/accounts";
  import { pollState } from "../lib/qrLogin";
  import QrDialog from "./QrDialog.svelte";

  // A scanned login code (see lib/qrLogin): the link opened, or the app's
  // scanner. Logged in, a logged-out device's code: "Log in on <device>?"
  // (the phishing warning: whoever shows it gets the account). Logged out,
  // a logged-in device's code: wait for its yes, then this device is
  // logged in (onLoggedIn). A logged-in device does not change accounts by
  // a code (the server refuses: 409).
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
    me: string | null; // the account logged in here (null: none, or not known yet)
    onLoggedIn: (name: string) => void;
    onClose: () => void;
  } = $props();

  let phase = $state<"scanning" | "ask" | "waiting" | "done" | "refused" | "expired" | "error">("scanning");
  let other = $state("");
  let error = $state("");
  let stop = () => {};
  onDestroy(() => stop());

  const failed = (err: unknown, scanning: boolean) => {
    const status = apiStatus(err);
    error =
      status === 404
        ? "This code is not valid any more. Make a new one."
        : status === 409
          ? "This code logs in a device without an account open. Log out here first, or scan it on that device."
          : status === 401 && scanning
            ? "Log in on this device first, then scan the code again."
            : err instanceof Error
              ? err.message
              : String(err);
    phase = "error";
  };
  // Once, as it opens (a code is scanned once).
  void untrack(() => api.qrScan(token, device)).then(
    (s) => {
      other = s.device;
      if (s.kind === "request") return void (phase = "ask");
      phase = "waiting";
      stop = pollState(
        () => api.qrWait(s.claim),
        (st) => {
          if (st.state === "done") (phase = "done"), onLoggedIn(st.name ?? "");
          else if (st.state === "refused" || st.state === "expired") phase = st.state;
          return phase === "waiting";
        },
      );
    },
    (err) => failed(err, true),
  );
  const answer = async (yes: boolean) => {
    try {
      await api.qrAnswer(token, yes);
      phase = yes ? "done" : "refused";
    } catch (err) {
      failed(err, false);
    }
  };
</script>

<QrDialog title="Log in with a QR code" {onClose}>
  {#if phase === "scanning"}
    <p class="note">Reading the code…</p>
  {:else if phase === "ask"}
    <p><strong>Log in on {other} {me ? `as ${me}` : "with your account"}?</strong></p>
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
  {:else}
    <p role="alert">{error}</p>
  {/if}
</QrDialog>
