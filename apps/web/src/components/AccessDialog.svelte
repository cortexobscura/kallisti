<script lang="ts">
  import type { ReaderGroup } from "../lib/members";

  // Who has access to a private channel, for its members (read-only; staff
  // and the channel's moderators change it in the Members dialog): its
  // members and the staff, online first (see readerGroups).
  let { name, groups, onClose }: { name: string; groups: readonly ReaderGroup<{ name: string }>[]; onClose: () => void } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  $effect(() => dialog?.showModal());
  const rowName = (r: ReaderGroup<{ name: string }>["rows"][number]) => ("user" in r ? r.user.name : "member" in r ? r.member.name : r.offline.name);
</script>

<dialog bind:this={dialog} onclose={onClose} aria-labelledby="access-dialog-title">
  <header>
    <h2 id="access-dialog-title">Members of {name}</h2>
    <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
  </header>
  <p class="note">Only members, admins and moderators see this channel.</p>
  {#each groups as g (g.title)}
    <h3>{g.title} — {g.rows.length}</h3>
    <ul aria-label={g.title}>
      {#each g.rows as r (rowName(r))}
        <li class:offline={"offline" in r}><span class="ellipsis">{rowName(r)}</span></li>
      {/each}
    </ul>
  {/each}
</dialog>

<style>
  dialog {
    width: min(440px, 92vw);
    max-height: 80vh;
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
  h3 {
    margin: 12px 0 6px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-muted);
  }
  .close {
    font-size: 24px;
    line-height: 1;
    color: var(--text-muted);
  }
  .note {
    margin: 0 0 4px;
    font-size: 14px;
    color: var(--text-muted);
  }
  ul {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  li {
    max-width: 100%;
    padding: 4px 10px;
    border-radius: 12px;
    background: var(--bg-rail);
    font-size: 14px;
  }
  li.offline {
    opacity: 0.55;
  }
</style>
