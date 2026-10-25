# Stream view: drag to resize

## Summary

The stream view takes a fixed part of the window height and covers much
of the chat. Users want to drag its lower edge up or down.

## Requirements

- A handle on the lower edge; dragging changes the height (mouse, touch,
  keyboard with the arrow keys).
- Limits so chat and stream both stay usable; the height is saved.

## Acceptance Criteria

- Dragging the edge resizes the view; it keeps its height after a reload.
- Tests for the limits and the saved value.
