<script lang="ts">
  import { initials } from "../lib/avatar";

  // A server in the rail: logo null shows its initials.
  export type RailServer = { id: string; name: string; logo: string | null; active: boolean };

  // Several servers in one client (see lib/serverList): a click shows one,
  // (+) adds one; without onAdd, no (+).
  let { servers, onSelect, onAdd }: { servers: readonly RailServer[]; onSelect?: (id: string) => void; onAdd?: () => void } = $props();
</script>

<nav class="rail" aria-label="Servers">
  {#each servers as s (s.id)}
    <div class="item">
      <span class="pill" class:active={s.active} aria-hidden="true"></span>
      <button class="server" class:active={s.active} title={s.name} aria-label={s.name} aria-current={s.active ? "true" : undefined} onclick={() => onSelect?.(s.id)}>
        {#if s.logo}<img src={s.logo} alt="" />{:else}{initials(s.name)}{/if}
      </button>
    </div>
  {/each}
  {#if onAdd}
    <div class="item">
      <button class="server add" title="Add a server" aria-label="Add a server" onclick={onAdd}>+</button>
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
  .add {
    color: var(--accent);
    font-size: 22px;
    font-weight: 400;
  }
  .server.active,
  .server:hover:not(:disabled) {
    background: var(--accent-strong);
    color: white;
  }
</style>
