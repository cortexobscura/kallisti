<script lang="ts">
  import { untrack } from "svelte";
  import { busyLabel } from "../lib/saveFeedback";
  import { createSave } from "../lib/saveFeedback.svelte";
  import type { Webhook } from "../lib/webhooks";
  import SaveStatus from "./SaveStatus.svelte";

  // A text channel's webhooks (staff, see lib/webhooks): the list, a new
  // one (its URL shown once), rename, delete. Each change applies at once.
  let {
    name,
    webhooks = null,
    created = null,
    onLoad = () => {},
    onCreate,
    onRename,
    onDelete,
    onClose,
  }: {
    name: string; // the channel's, as shown
    webhooks?: readonly Webhook[] | null; // null: not loaded yet
    created?: { id: number; url: string } | null; // the new one's URL
    onLoad?: () => void;
    onCreate: (name: string) => Promise<void>;
    onRename: (id: number, name: string) => Promise<void>;
    onDelete: (id: number) => Promise<void>;
    onClose: () => void;
  } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  $effect(() => dialog?.showModal());
  $effect(() => untrack(onLoad));

  let newName = $state("");
  let copied = $state(false);
  const copy = (url: string) => void navigator.clipboard?.writeText(url).then(() => (copied = true));
  // The one being renamed (its new name), the one asked to delete.
  let renaming = $state<{ id: number; name: string } | null>(null);
  let deleting = $state<number | null>(null);
  const changes = createSave();
  const create = (event: SubmitEvent) => {
    event.preventDefault();
    const n = newName.trim();
    if (!n) return;
    copied = false;
    changes.run(() => onCreate(n).then(() => (newName = "")));
  };
  const rename = (event: SubmitEvent) => {
    event.preventDefault();
    const r = renaming;
    if (r?.name.trim()) changes.run(() => onRename(r.id, r.name.trim()).then(() => (renaming = null)));
  };
  const when = (at: number) => new Date(at).toLocaleDateString();
</script>

<dialog bind:this={dialog} onclose={onClose} aria-labelledby="webhooks-dialog-title">
  <header>
    <h2 id="webhooks-dialog-title">Webhooks of {name}</h2>
    <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
  </header>
  <p class="note">
    A webhook URL posts messages into this channel. Tools with a Discord webhook setting (Gitea, GitHub Actions, Grafana, Uptime Kuma and others) can use it as it
    is.
  </p>
  <!-- Until it is deleted. -->
  {#if created && (webhooks === null || webhooks.some((w) => w.id === created.id))}
    <div class="row">
      <input readonly value={created.url} aria-label="New webhook URL" onfocus={(e) => e.currentTarget.select()} />
      <button type="button" class="action" onclick={() => copy(created.url)}>{copied ? "Copied" : "Copy"}</button>
    </div>
    <p class="note">Copy it now: it is shown only once. Keep it secret: everyone with it can post here.</p>
  {/if}
  <form class="row" onsubmit={create}>
    <input bind:value={newName} maxlength="80" placeholder="Name, e.g. CI" aria-label="Name of a new webhook" />
    <button type="submit" class="action" disabled={changes.busy || !newName.trim()}>{busyLabel(changes.state, "Create", "Working…")}</button>
  </form>
  <SaveStatus save={changes.state} />
  {#if webhooks === null}
    <p class="note">Loading webhooks…</p>
  {:else if webhooks.length}
    <ul aria-label="Webhooks">
      {#each webhooks as w (w.id)}
        <li>
          {#if renaming?.id === w.id}
            <form class="row" onsubmit={rename}>
              <!-- svelte-ignore a11y_autofocus -->
              <input bind:value={renaming.name} maxlength="80" aria-label="New name of {w.name}" autofocus />
              <button type="submit" class="action" disabled={changes.busy}>Save</button>
              <button type="button" class="action" onclick={() => (renaming = null)}>Cancel</button>
            </form>
          {:else}
            <span class="what">
              <span class="ellipsis">{w.name}</span>
              <span class="sub">by {w.creator}, {when(w.created)}</span>
            </span>
            {#if deleting === w.id}
              <button type="button" class="danger" disabled={changes.busy} onclick={() => changes.run(() => onDelete(w.id).then(() => (deleting = null)))}>
                Delete it
              </button>
              <button type="button" onclick={() => (deleting = null)}>Keep</button>
            {:else}
              <button type="button" onclick={() => ((renaming = { id: w.id, name: w.name }), (deleting = null))}>Rename</button>
              <button type="button" onclick={() => ((deleting = w.id), (renaming = null))} aria-label="Delete {w.name}">Delete</button>
            {/if}
          {/if}
        </li>
      {/each}
    </ul>
  {:else}
    <p class="note">No webhooks yet.</p>
  {/if}
</dialog>

<style>
  dialog {
    width: min(480px, 92vw);
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
  .row {
    display: flex;
    gap: 8px;
    margin-bottom: 8px;
  }
  .row > * {
    min-width: 0;
  }
  .row input {
    flex: 1;
    padding: 10px;
    border: none;
    border-radius: 4px;
    background: var(--bg-rail);
    color: var(--text);
    font: inherit;
  }
  .action {
    flex: none;
    padding: 10px 12px;
    border-radius: 4px;
    background: var(--bg-rail);
    color: var(--text);
  }
  ul {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 12px 0 0;
    padding: 0;
    list-style: none;
  }
  li {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 6px 6px 10px;
    border-radius: 4px;
    background: var(--bg-rail);
    font-size: 14px;
  }
  li form {
    flex: 1;
    margin: 0;
  }
  li .action,
  li input {
    background: var(--bg-main);
  }
  .what {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }
  .sub {
    font-size: 12px;
    color: var(--text-muted);
  }
  li > button {
    padding: 4px 8px;
    border-radius: 4px;
    color: var(--text-muted);
  }
  li > button:hover {
    color: var(--text);
  }
  li > .danger {
    color: var(--danger);
  }
</style>
