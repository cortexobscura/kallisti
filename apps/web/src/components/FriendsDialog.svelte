<script lang="ts">
  import type { FriendRow } from "../lib/friends";

  // The friends list (see lib/friends): online ones first, where they are
  // (only the servers in the rail tell). Message opens a direct chat on
  // that server (undefined: the server is not in the rail, or no login
  // there). Friends are added from a user's menu.
  let {
    rows,
    canMessage,
    onMessage,
    onRemove,
    onClose,
  }: {
    rows: readonly FriendRow[];
    canMessage: (origin: string) => boolean;
    onMessage: (origin: string, user: number) => void;
    onRemove: (origin: string, user: number) => void;
    onClose: () => void;
  } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  $effect(() => dialog?.showModal());
  const host = (origin: string) => new URL(origin).host;
  const message = (origin: string, user: number) => {
    onMessage(origin, user);
    dialog?.close();
  };
</script>

<dialog bind:this={dialog} onclose={onClose} aria-labelledby="friends-dialog-title">
  <header>
    <h2 id="friends-dialog-title">Friends</h2>
    <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
  </header>
  {#if rows.length}
    <ul>
      {#each rows as r (`${r.friend.origin} ${r.friend.user}`)}
        <li class:offline={!r.online}>
          <span class="dot" class:on={r.online} aria-hidden="true"></span>
          <span class="who">
            <span class="name ellipsis">{r.friend.name}</span>
            <span class="where ellipsis">
              {r.server ?? host(r.friend.origin)}{#if r.online}{r.voice ? ` · in ${r.voice}` : " · online"}{:else} · offline{/if}
            </span>
          </span>
          {#if canMessage(r.friend.origin)}
            <button type="button" class="secondary" aria-label="Message {r.friend.name}" onclick={() => message(r.friend.origin, r.friend.user)}>Message</button>
          {/if}
          <button type="button" class="secondary" aria-label="Remove {r.friend.name} from your friends" onclick={() => onRemove(r.friend.origin, r.friend.user)}>Remove</button>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="note">No friends yet. Add someone from their menu (click their name).</p>
  {/if}
  <p class="note">Only the servers in your list tell who is online.</p>
</dialog>

<style>
  dialog {
    width: min(460px, 92vw);
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
  .close {
    font-size: 24px;
    line-height: 1;
    color: var(--text-muted);
  }
  .note {
    margin: 12px 0 0;
    font-size: 13px;
    color: var(--text-muted);
  }
  ul {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  li {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  li.offline .who {
    opacity: 0.6;
  }
  .dot {
    flex: none;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--text-faint);
  }
  .dot.on {
    background: var(--online);
  }
  .who {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }
  .name {
    font-weight: 600;
  }
  .where {
    font-size: 12px;
    color: var(--text-muted);
  }
  .secondary {
    flex: none;
    padding: 6px 12px;
    border-radius: 6px;
    background: var(--bg-input);
    font-weight: 600;
  }
</style>
