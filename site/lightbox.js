// Screenshots open at full size on click, Enter or Space (the files are
// twice the size shown). A click anywhere, Escape or the button closes
// the picture; focus goes back to it. Without JavaScript the pictures
// stay as they are.
const box = document.createElement("dialog");
box.className = "lightbox";
box.innerHTML = '<button type="button" class="lightbox-close" aria-label="Close">×</button><img alt="" /><p class="lightbox-caption"></p>';
document.body.append(box);
const big = box.querySelector("img");
const caption = box.querySelector("p");
let opener = null;

const open = (img) => {
  opener = img;
  big.src = img.currentSrc || img.src;
  big.alt = img.alt;
  caption.textContent = img.alt;
  document.documentElement.classList.add("lightbox-open");
  box.showModal();
};

box.addEventListener("click", () => box.close());
box.addEventListener("close", () => {
  document.documentElement.classList.remove("lightbox-open");
  opener?.focus();
});

for (const img of document.querySelectorAll(".feature img, .hero-shot img")) {
  img.classList.add("zoomable");
  img.tabIndex = 0;
  img.setAttribute("role", "button");
  img.setAttribute("aria-label", `Enlarge: ${img.alt}`);
  img.addEventListener("click", () => open(img));
  img.addEventListener("keydown", (e) => {
    if (e.key !== "Enter" && e.key !== " ") return;
    e.preventDefault();
    open(img);
  });
}
