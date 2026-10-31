<script lang="ts">
  import { onDestroy } from "svelte";
  import type { QrStart, QrState } from "../lib/accounts";
  import { pollState, qrLink } from "../lib/qrLogin";
  import QrCode from "./QrCode.svelte";
  import QrDialog from "./QrDialog.svelte";

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

  let code = $state<QrStart | null>(null);
  let phase = $state<"loading" | "showing" | "ask" | "confirm" | "approving" | "done" | "refused" | "expired" | "error">("loading");
  let other = $state(""); // the other device
  let error = $state("");
  let stop = () => {};
  onDestroy(() => stop());

  // Its state, until it ends.
  const follow = (c: QrStart) =>
    pollState(
      () => api.qrWait(c.owner),
      (st) => {
        if (st.device) other = st.device;
        if (st.state === "scanned" && phase === "showing") phase = c.kind === "offer" ? "ask" : "confirm";
        else if (st.state === "approved" && c.kind === "offer") phase = "approving";
        else if (st.state === "done") {
          phase = "done"; // an offer: the phone has its session; a request: this device
          if (c.kind === "request") onLoggedIn?.(st.name ?? "");
        } else if (st.state === "refused" || st.state === "expired") phase = st.state;
        return phase !== "done" && phase !== "refused" && phase !== "expired";
      },
    );
  const failed = (err: unknown) => {
    error = err instanceof Error ? err.message : String(err);
    phase = "error";
  };
  const start = async () => {
    stop();
    phase = "loading";
    other = "";
    try {
      code = await api.qrStart(device);
      phase = "showing";
      stop = follow(code);
    } catch (err) {
      failed(err);
    }
  };
  void start();
  const answer = async (yes: boolean) => {
    if (!code) return;
    try {
      await api.qrAnswer(code.owner, yes);
      phase = yes ? "approving" : "refused";
    } catch (err) {
      failed(err);
    }
  };
</script>

<QrDialog title={code?.kind === "request" ? "Log in with a QR code" : "Log in on another device"} {onClose}>
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
  {:else if phase === "refused" || phase === "expired"}
    <p>{phase === "refused" ? "Not logged in: the login was refused." : "The code expired."}</p>
    <div class="buttons"><button type="button" class="primary" onclick={start}>New code</button></div>
  {:else}
    <p role="alert">{error}</p>
    <div class="buttons"><button type="button" class="primary" onclick={start}>Try again</button></div>
  {/if}
</QrDialog>
