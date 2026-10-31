<script lang="ts">
  import { initials } from "../lib/avatar";
  import Icon from "./Icon.svelte";

  // A server in the rail: logo null shows its initials; mentions (a red
  // count) and unread (a dot) when it is not on screen.
  export type RailServer = { id: string; name: string; logo: string | null; active: boolean; mentions?: number; unread?: boolean };

  // Several servers in one client (see lib/serverList): a click shows one,
  // (+) adds one; without onAdd, no (+).
  // onFriends: the friends list (see lib/friends), above the servers.
  let {
    servers,
    onSelect,
    onAdd,
    onFriends,
  }: { servers: readonly RailServer[]; onSelect?: (id: string) => void; onAdd?: () => void; onFriends?: () => void } = $props();
</script>

<nav class="rail" aria-label="Servers">
  {#if onFriends}
    <div class="item">
      <button class="server friends" title="Friends" aria-label="Friends" onclick={onFriends}><Icon name="friends" size={26} /></button>
    </div>
    <div class="sep" aria-hidden="true"></div>
  {/if}
  {#each servers as s (s.id)}
    {@const mentions = s.active ? 0 : (s.mentions ?? 0)}
    <div class="item">
      <span class="pill" class:active={s.active} class:unread={!s.active && s.unread} aria-hidden="true"></span>
      <button
        class="server"
        class:active={s.active}
        title={s.name}
        aria-label={mentions ? `${s.name}, ${mentions} unread ${mentions === 1 ? "mention or direct message" : "mentions or direct messages"}` : !s.active && s.unread ? `${s.name}, unread messages` : s.name}
        aria-current={s.active ? "true" : undefined}
        onclick={() => onSelect?.(s.id)}
      >
        {#if s.logo}<img src={s.logo} alt="" />{:else}{initials(s.name)}{/if}
      </button>
      {#if mentions}<span class="badge" aria-hidden="true">{mentions > 99 ? "99+" : mentions}</span>{/if}
    </div>
  {/each}
  {#if onAdd}
    <div class="item">
      <button class="server add" title="Add a server" aria-label="Add a server" onclick={onAdd}><Icon name="plus" size={28} /></button>
    </div>
  {/if}
</nav>

<style>
  .rail {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 12px 0;
    background: var(--bg-rail);
  }
  .sep {
    width: 32px;
    height: 2px;
    border-radius: 1px;
    background: var(--border-subtle);
  }
  .item {
    position: relative;
    display: flex;
    justify-content: center;
    width: 100%;
  }
  /* The white pill on the left edge marks the current server. */
  .pill {
    position: absolute;
    top: 50%;
    left: 0;
    width: 4px;
    height: 8px;
    border-radius: 0 4px 4px 0;
    background: var(--text);
    opacity: 0;
    transform: translateY(-50%);
    transition:
      height 0.15s,
      opacity 0.15s;
  }
  .item:hover .pill {
    height: 20px;
    opacity: 1;
  }
  .pill.active {
    height: 40px;
    opacity: 1;
  }
  .pill.unread {
    opacity: 1;
  }
  /* The red count of mentions, at the icon's lower right. */
  .badge {
    position: absolute;
    right: 12px;
    bottom: -4px;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border: 3px solid var(--bg-rail);
    border-radius: 8px;
    box-sizing: content-box;
    background: #c62f32; /* in both themes: white on it is readable */
    color: white;
    font-size: 11px;
    font-weight: 700;
    line-height: 16px;
    text-align: center;
    pointer-events: none;
  }
  .server {
    width: 40px;
    height: 40px;
    padding: 0;
    border-radius: 12px;
    background: var(--bg-input);
    font-size: 16px;
    font-weight: 600;
    transition:
      background 0.15s,
      color 0.15s;
  }
  .server img {
    display: block;
    width: 100%;
    height: 100%;
    border-radius: inherit;
    object-fit: cover;
  }
  /* An icon, not a "+": a character's height depends on the font. */
  .add,
  .friends {
    display: grid;
    place-items: center;
    color: var(--accent);
  }
  .server.active,
  .server:hover:not(:disabled) {
    background: var(--accent-strong);
    color: white;
  }
</style>
