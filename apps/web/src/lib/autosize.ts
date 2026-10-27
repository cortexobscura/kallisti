// A Svelte action: a textarea as tall as its text (CSS max-height caps it,
// then it scrolls). Pass the value, so it fits again when it changes; after
// the update, when the new value is in the element.
// Not while hidden (a server's client out of sight: no size to measure);
// again when it gets the focus.
export const autosize = (node: HTMLTextAreaElement, _value?: string) => {
  const fit = () => {
    if (node.getClientRects().length === 0) return;
    node.style.height = "auto";
    node.style.height = `${node.scrollHeight}px`;
  };
  fit();
  node.addEventListener("focus", fit);
  return { update: () => queueMicrotask(fit), destroy: () => node.removeEventListener("focus", fit) };
};
