<script lang="ts">
  import { qrPath } from "../lib/qrLogin";

  // A QR code, dark on light whatever the theme (scanners want it so).
  let { text, label, size = 220 }: { text: string; label: string; size?: number } = $props();
  const code = $derived(qrPath(text));
</script>

<svg width={size} height={size} viewBox="0 0 {code.size} {code.size}" shape-rendering="crispEdges" role="img" aria-label={label}>
  <rect width={code.size} height={code.size} fill="#fff" />
  <path d={code.path} fill="#000" />
</svg>

<style>
  svg {
    display: block;
    border-radius: 8px;
  }
</style>
