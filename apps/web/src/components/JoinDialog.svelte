<script lang="ts">
  // Before a link (or an invite) joins a voice channel: asked first, and
  // told when it ends a call.
  let { name, inCall, onJoin, onClose }: { name: string; inCall: boolean; onJoin: () => void; onClose: () => void } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  $effect(() => dialog?.showModal());
  const join = () => {
    onJoin();
    dialog?.close();
  };
</script>

<dialog bind:this={dialog} onclose={onClose} aria-labelledby="join-dialog-title">
  <h2 id="join-dialog-title">Join {name}?</h2>
  <p>A link leads to this voice channel. Others there hear you once you join.</p>
  {#if inCall}<p class="note">This ends your current call.</p>{/if}
  <footer>
    <button type="button" class="secondary" onclick={() => dialog?.close()}>Cancel</button>
    <!-- svelte-ignore a11y_autofocus -->
    <button type="button" class="primary" onclick={join} autofocus>Join</button>
  </footer>
</dialog>

<style>
  dialog {
    width: min(400px, 92vw);
    padding: 24px;
    border: none;
    border-radius: 8px;
    background: var(--bg-main);
    color: var(--text);
  }
  dialog::backdrop {
    background: rgb(0 0 0 / 0.6);
  }
  h2 {
    margin: 0 0 12px;
    font-size: 18px;
    overflow-wrap: anywhere;
  }
  p {
    margin: 0 0 12px;
    font-size: 14px;
  }
  .note {
    color: var(--text-muted);
  }
  footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 16px;
  }
  footer button {
    padding: 8px 16px;
    border-radius: 6px;
    font-weight: 600;
  }
  .secondary {
    background: var(--bg-input);
  }
  .primary {
    background: var(--accent-strong);
    color: white;
  }
</style>
