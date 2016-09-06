//
// Main: all custom scripting should go here.
//

// The equivalent of JQuery's `$(document).ready`. It's a little newer, but
// caniusethis estimates a 97+% penetration rate which is good enough for me.
window.addEventListener('DOMContentLoaded', function () {
  hljs.initHighlightingOnLoad();
})

// Highlight.js wants to be in `DOMContentLoaded` (and not `load`) while
// Retina.js wants the opposite. I don't quite understand why this is, but I'm
// cargo culting for now.
window.addEventListener('load', function () {
  // With no arguments looks for anything with a `data-rjs` tag.
  retinajs();
});
