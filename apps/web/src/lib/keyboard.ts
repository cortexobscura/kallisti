// Is the on-screen keyboard open (phones)? Browsers do not tell; the visual
// viewport gets much shorter than the layout while it is open. Then the
// home indicator space below the composer only makes a gap above the
// keyboard (app.css: html[data-keyboard]).

const MIN_KEYBOARD = 150; // px: less is browser chrome, not a keyboard

export const keyboardOpen = ({ layout, visual, scale }: { layout: number; visual: number; scale: number }) =>
  layout - visual * scale > MIN_KEYBOARD;

// The parts of window.visualViewport used here.
type Viewport = {
  height: number;
  scale: number;
  addEventListener(type: "resize", cb: () => void): void;
  removeEventListener(type: "resize", cb: () => void): void;
};

// onChange gets each change of the keyboard state. Returns a stop function.
export const watchKeyboard = (onChange: (open: boolean) => void, { viewport, layoutHeight }: { viewport: Viewport; layoutHeight: () => number }) => {
  let open = false;
  const update = () => {
    const now = keyboardOpen({ layout: layoutHeight(), visual: viewport.height, scale: viewport.scale });
    if (now !== open) onChange((open = now));
  };
  viewport.addEventListener("resize", update);
  update();
  return () => viewport.removeEventListener("resize", update);
};
