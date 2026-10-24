<script lang="ts">
  import { EMOJI_TYPES, emojiNameFromFile } from "../lib/customEmoji";
  import { busyLabel } from "../lib/saveFeedback";
  import { createSave } from "../lib/saveFeedback.svelte";
  import { MAX_STICKER_BYTES, STICKER_SIZE, stickerForm, type Sticker, type StickerFields } from "../lib/stickers";
  import EmojiImage from "./EmojiImage.svelte";
  import SaveStatus from "./SaveStatus.svelte";

  // The server's stickers (moderation dialog): add, edit (name,
  // description, related emoji, pack) and delete. url: absolute.
  let {
    list,
    onAdd,
    onEdit,
    onDelete,
  }: {
    list: readonly Sticker[];
    onAdd: (fields: StickerFields, file: File) => Promise<void>;
    onEdit: (id: number, fields: StickerFields) => Promise<void>;
    onDelete: (id: number) => Promise<void>;
  } = $props();

  const blank = (): StickerFields => ({ name: "", description: "", emoji: "", pack: "" });
  const trimmed = (f: StickerFields): StickerFields => ({ name: f.name.trim(), description: f.description.trim(), emoji: f.emoji.trim(), pack: f.pack.trim() });
  const add = createSave();
  const change = createSave();
  let fields = $state(blank());
  let file = $state<File | null>(null);
  let editing = $state<({ id: number } & StickerFields) | null>(null);
  let confirming = $state<number | null>(null);
  let fileInput: HTMLInputElement | undefined = $state();
  let image: EmojiImage | undefined = $state();
  let done = $state("Deleted");

  const submit = (e: SubmitEvent) => {
    e.preventDefault();
    const f = trimmed(fields);
    add.run(async () => {
      const problem = stickerForm(f, list);
      if (problem) throw new Error(problem);
      // Checked as uploaded: still images are scaled down first.
      const out = (await image?.output()) ?? null;
      if (!out) throw new Error("Choose an image.");
      if (out.size > MAX_STICKER_BYTES) throw new Error("A sticker has at most 512 KB.");
      await onAdd(f, out);
      fields = blank();
      file = null;
      if (fileInput) fileInput.value = "";
    });
  };
  const save = (e: SubmitEvent) => {
    e.preventDefault();
    if (!editing) return;
    const { id, ...rest } = editing;
    const f = trimmed(rest);
    const problem = stickerForm(f, list, id);
    done = "Saved";
    change.run(async () => {
      if (problem) throw new Error(problem);
      await onEdit(id, f);
      editing = null;
    });
  };
</script>

{#snippet fieldInputs(f: StickerFields, what: string)}
  <input bind:value={f.name} placeholder="name" aria-label="{what} name" maxlength="32" />
  <input bind:value={f.emoji} placeholder="emoji" aria-label="{what} related emoji" class="short" />
  <input bind:value={f.pack} placeholder="pack" aria-label="{what} pack" maxlength="32" />
  <input bind:value={f.description} placeholder="description or credit" aria-label="{what} description" maxlength="100" class="wide" />
{/snippet}

<form class="fields" onsubmit={submit}>
  <input
    bind:this={fileInput}
    type="file"
    disabled={add.busy}
    accept={EMOJI_TYPES.join(",")}
    aria-label="Sticker image"
    onchange={(e) => {
      file = e.currentTarget.files?.[0] ?? null;
      if (file && !fields.name) fields.name = emojiNameFromFile(file.name);
    }}
  />
  {@render fieldInputs(fields, "Sticker")}
  <button type="submit" disabled={add.busy}>{busyLabel(add.state, "Add", "Adding…")}</button>
  <SaveStatus save={add.state} done="Added" />
</form>
<p class="note">
  PNG, GIF, WebP or JPEG, at most 512 KB; still images are scaled to fit {STICKER_SIZE}×{STICKER_SIZE} px. The emoji and pack are
  optional: the picker groups by pack and finds stickers by their emoji. Put a licence's credit in the description.
</p>
<EmojiImage bind:this={image} {file} name={fields.name.trim()} max={STICKER_SIZE} limit="512 KB" />
{#if list.length === 0}
  <p class="note">This server has no stickers yet.</p>
{:else}
  <ul>
    {#each list as s (s.id)}
      <li>
        <img class="sticker-image" src={s.url} alt="" />
        {#if editing?.id === s.id}
          <form class="fields" onsubmit={save}>
            {@render fieldInputs(editing, s.name)}
            <button type="submit" disabled={change.busy}>Save</button>
            <button type="button" onclick={() => (editing = null)}>Cancel</button>
          </form>
        {:else}
          <div class="what">
            <strong>{s.name}</strong>{#if s.emoji}&nbsp;{s.emoji}{/if}{#if s.pack}<span class="pack"> · {s.pack}</span>{/if}
            {#if s.description}<div class="description">{s.description}</div>{/if}
          </div>
          {#if confirming === s.id}
            <button
              type="button"
              class="danger"
              disabled={change.busy}
              onclick={() => ((done = "Deleted"), change.run(() => onDelete(s.id)), (confirming = null))}
            >
              Delete
            </button>
            <button type="button" onclick={() => (confirming = null)}>Cancel</button>
          {:else}
            <button
              type="button"
              onclick={() => (editing = { id: s.id, name: s.name, description: s.description ?? "", emoji: s.emoji ?? "", pack: s.pack ?? "" })}
              >Edit</button
            >
            <button type="button" onclick={() => (confirming = s.id)}>Delete…</button>
          {/if}
        {/if}
      </li>
    {/each}
  </ul>
{/if}
<SaveStatus save={change.state} {done} />

<style>
  .fields {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    gap: 6px;
    align-items: center;
  }
  input:not([type="file"]) {
    flex: 1;
    min-width: 6em;
    padding: 6px 8px;
    border: none;
    border-radius: 4px;
    background: var(--bg-rail);
    color: var(--text);
    font: inherit;
  }
  input.short {
    flex: 0 1 5em;
    min-width: 4em;
  }
  input.wide {
    flex-basis: 100%;
  }
  input[type="file"] {
    max-width: 100%;
  }
  button {
    padding: 6px 10px;
    border-radius: 4px;
    background: var(--bg-input);
    color: var(--text);
  }
  .danger {
    background: var(--danger);
    color: white;
  }
  .note {
    color: var(--text-muted);
    font-size: 14px;
  }
  ul {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  li {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 0;
    border-bottom: 1px solid var(--border-subtle);
  }
  .what {
    flex: 1;
    min-width: 0;
  }
  .pack,
  .description {
    color: var(--text-muted);
    font-size: 13px;
  }
  .sticker-image {
    width: 64px;
    height: 64px;
    object-fit: contain;
  }
</style>
