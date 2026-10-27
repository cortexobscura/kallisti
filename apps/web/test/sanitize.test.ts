// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { sanitize } from "../src/lib/sanitize";

const img = (alt: string, code: string) => `<img class="emoji" draggable="false" alt="${alt}" src="/emoji/${code}.png">`;

describe("sanitize with emoji", () => {
  it("makes emoji in text Twemoji images, after sanitizing", () => {
    expect(sanitize("hi 😀 <b>❤️</b>", true)).toBe(`hi ${img("😀", "1f600")} <b>${img("❤️", "2764")}</b>`);
  });
  it("never touches attributes, and attribute text cannot become markup", () => {
    const div = document.createElement("div");
    div.innerHTML = sanitize('<a title="😀 <img src=x onerror=alert(1)>" href="https://x.org/">hi</a>', true);
    expect(div.querySelectorAll("img").length).toBe(0); // the title stays text
    expect(div.querySelector("a")?.title).toBe("😀 <img src=x onerror=alert(1)>");
  });
  it("makes images in code too (blank there without an emoji font)", () => {
    expect(sanitize("<code>😀</code>", true)).toBe(`<code>${img("😀", "1f600")}</code>`);
  });
  it("does not let foreign images in, also with the emoji class", () => {
    expect(sanitize('<img class="emoji" src="/emoji/../api/x.png">', true)).not.toContain("src=");
  });
  it("leaves emoji alone without the option", () => {
    expect(sanitize("hi 😀")).toBe("hi 😀");
  });
});

// The page may load pictures and media from other servers (several servers
// in one client): message HTML must load nothing remote (readers' IPs).
describe("sanitize loads nothing remote", () => {
  const remote = (html: string) => {
    const out = sanitize(html);
    return /tracker\.example/.test(out) ? out : null;
  };
  it("drops remote addresses from styles, srcset, posters, SVG and media", () => {
    for (const html of [
      '<span style="background-image:url(https://tracker.example/a)">x</span>',
      '<span style="background:\\75 rl(https://tracker.example/a)">x</span>',
      '<span style="background-image:image-set(\'https://tracker.example/a\' 1x)">x</span>',
      '<img srcset="https://tracker.example/a 1x" src="data:image/png;base64,AA==">',
      '<video poster="https://tracker.example/a"></video>',
      '<audio src="https://tracker.example/a"></audio>',
      '<video><source src="https://tracker.example/a"></video>',
      '<svg><image href="https://tracker.example/a"></image></svg>',
      '<svg><image xlink:href="https://tracker.example/a"></image></svg>',
      '<table background="https://tracker.example/a"><tr><td>x</td></tr></table>',
      '<p>a</p><style>p{background:url(https://tracker.example/a)}</style>',
      '<svg><style>circle{fill:url(https://tracker.example/a)}</style><circle r="5"/></svg>',
    ]) {
      expect(remote(html), html).toBeNull();
    }
  });
  it("keeps colours, links and inline pictures", () => {
    expect(sanitize('<span style="color: red">x</span>')).toBe('<span style="color: red">x</span>');
    expect(sanitize("<span style=\"font-family: 'Curlz MT'; color: red\">x</span>")).toContain("color: red");
    expect(sanitize('<a href="https://x.org/">x</a>')).toContain('href="https://x.org/"');
    expect(sanitize('<img src="data:image/png;base64,AA==">')).toContain("data:image/png");
  });
});
