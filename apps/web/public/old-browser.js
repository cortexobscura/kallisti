// Browsers too old for the app's code (a WebView never updated on an old
// Android) skip or fail its module script: an empty page. Plain old syntax
// here, so they run this and tell why. toSorted (Chrome 110, Safari 16,
// Firefox 115): used by the app's code, which needs no newer browser.
if (typeof Array.prototype.toSorted !== "function") {
  (function () {
    var app = document.getElementById("app");
    var box = document.createElement("div");
    box.style.maxWidth = "360px";
    box.style.margin = "64px auto";
    box.style.padding = "0 16px";
    box.style.font = "16px/1.5 sans-serif";
    box.style.color = "#ddd";
    var line = function (text, bold) {
      var p = document.createElement("p");
      p.textContent = text;
      if (bold) p.style.fontWeight = "bold";
      box.appendChild(p);
    };
    line("This browser is too old for Kallisti.", true);
    // "wv": the user agent of Android's WebView (the Android app).
    if (/; wv\)/.test(navigator.userAgent))
      line("Update “Android System WebView” (or Chrome) in the Play Store, then open the app again.");
    else line("Update your browser, or open Kallisti in a newer one.");
    app.appendChild(box);
    // A browser a bit newer runs the app's module too: with no #app it
    // does not mount below the message.
    app.removeAttribute("id");
  })();
}
