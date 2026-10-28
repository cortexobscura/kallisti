<script lang="ts">
  import { splitName } from "../lib/userName";

  // A name; an account from another server's home (@host) faded.
  let { name }: { name: string } = $props();
  const parts = $derived(splitName(name));
</script>

{parts.own}{#if parts.host}<span class="host">@{parts.host}</span>{/if}

<style>
  .host {
    opacity: 0.55;
    font-weight: normal;
  }
</style>
