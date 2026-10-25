<script lang="ts">
  import { tick } from "svelte";
  import { emptyPollDraft, pollBody, pollProblem, POLL_ANSWER_MAX, POLL_ANSWERS_MAX, POLL_DURATIONS, POLL_QUESTION_MAX, type PollBody } from "../lib/polls";

  // A new poll (see lib/polls): question, answers with an emoji each,
  // duration, multiple choice, anonymous votes. onSubmit settles once the
  // server stored it (or fails with why: the form stays).
  let { onSubmit, onClose }: { onSubmit: (poll: PollBody) => Promise<void>; onClose: () => void } = $props();

  let dialog: HTMLDialogElement | undefined = $state();
  let draft = $state(emptyPollDraft());
  let error = $state("");
  let sending = $state(false);
  $effect(() => dialog?.showModal());
  const problem = $derived(pollProblem(draft));
  const addAnswer = async () => {
    draft.answers.push({ text: "" });
    await tick();
    dialog?.querySelector<HTMLInputElement>(".answer:last-of-type input.text")?.focus();
  };
  const submit = async (event: SubmitEvent) => {
    event.preventDefault();
    if (problem || sending) return void (error = problem ?? "");
    sending = true;
    error = "";
    try {
      await onSubmit(pollBody($state.snapshot(draft)));
      dialog?.close();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      sending = false;
    }
  };
</script>

<!-- Not closed while posting: a refusal shows here. -->
<dialog bind:this={dialog} onclose={onClose} oncancel={(e) => sending && e.preventDefault()} aria-labelledby="poll-title">
  <form onsubmit={submit}>
    <header>
      <h2 id="poll-title">Create a poll</h2>
      <button type="button" class="close" onclick={() => dialog?.close()} aria-label="Close">×</button>
    </header>
    <label>
      Question
      <!-- svelte-ignore a11y_autofocus -->
      <input bind:value={draft.question} maxlength={POLL_QUESTION_MAX} placeholder="What do you want to ask?" required autofocus />
    </label>
    <fieldset>
      <legend>Answers</legend>
      {#each draft.answers as answer, i (i)}
        <div class="answer">
          <input class="emoji" bind:value={answer.emoji} maxlength="16" placeholder="🙂" aria-label="Emoji of answer {i + 1} (optional)" />
          <input class="text" bind:value={answer.text} maxlength={POLL_ANSWER_MAX} placeholder="Answer {i + 1}" aria-label="Answer {i + 1}" />
          {#if draft.answers.length > 2}
            <button type="button" class="remove" onclick={() => draft.answers.splice(i, 1)} aria-label="Remove answer {i + 1}">×</button>
          {/if}
        </div>
      {/each}
      {#if draft.answers.length < POLL_ANSWERS_MAX}
        <button type="button" class="add" onclick={addAnswer}>+ Add another answer</button>
      {/if}
    </fieldset>
    <label>
      Duration
      <select bind:value={draft.hours}>
        {#each POLL_DURATIONS as [hours, text] (hours)}
          <option value={hours}>{text}</option>
        {/each}
      </select>
    </label>
    <label class="check"><input type="checkbox" bind:checked={draft.multiple} /> Allow more than one answer</label>
    <label class="check"><input type="checkbox" bind:checked={draft.anonymous} /> Anonymous votes (nobody sees who voted for what)</label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <footer>
      <button type="button" class="secondary" onclick={() => dialog?.close()}>Cancel</button>
      <button type="submit" class="primary" disabled={!!problem || sending}>{sending ? "Posting…" : "Post"}</button>
    </footer>
  </form>
</dialog>

<style>
  dialog {
    width: min(480px, 92vw);
    max-height: 90vh;
    padding: 24px;
    border: none;
    border-radius: 8px;
    background: var(--bg-main);
    color: var(--text);
  }
  dialog::backdrop {
    background: rgb(0 0 0 / 0.6);
  }
  header,
  footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  header {
    padding-bottom: 16px;
  }
  footer {
    justify-content: flex-end;
    gap: 8px;
    margin-top: 8px;
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
  label,
  legend {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-muted);
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 0 0 16px;
  }
  label.check {
    flex-direction: row;
    align-items: center;
    gap: 8px;
    font-weight: 400;
    text-transform: none;
    color: var(--text);
    font-size: 14px;
  }
  fieldset {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 0 0 16px;
    padding: 0;
    border: none;
  }
  legend {
    margin-bottom: 6px;
    padding: 0;
  }
  .answer {
    display: flex;
    gap: 6px;
  }
  input:not([type="checkbox"]),
  select {
    padding: 10px;
    border: none;
    border-radius: 4px;
    background: var(--bg-rail);
    color: var(--text);
    font: inherit;
    font-size: 14px; /* not the label's */
    font-weight: 400;
    text-transform: none;
  }
  input.emoji {
    width: 3.5em;
    text-align: center;
  }
  input.text {
    flex: 1;
    min-width: 0;
  }
  .remove {
    padding: 0 8px;
    font-size: 20px;
    color: var(--text-muted);
  }
  .add {
    align-self: flex-start;
    padding: 4px 0;
    color: var(--link);
  }
  .error {
    margin: 0 0 8px;
    color: var(--danger);
  }
  footer button {
    padding: 10px 16px;
    border-radius: 4px;
    color: white;
    font-weight: 600;
  }
  footer .secondary {
    background: none;
    color: var(--text);
  }
  footer .primary {
    background: var(--accent-strong);
  }
  footer .primary:disabled {
    opacity: 0.5;
  }
</style>
