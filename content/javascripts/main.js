//
// Main: all custom scripting should go here.
//

// The equivalent of JQuery's `$(document).ready`. It's a little newer, but
// caniusethis estimates a 97+% penetration rate which is good enough for me.
window.addEventListener('DOMContentLoaded', function () {
  hljs.initHighlightingOnLoad();
})

// This is for the photos page. It uses intersection observer (a browser
// built-in feature) to load images as they're scrolled to so that we save
// bandwidth.
document.addEventListener("DOMContentLoaded", function() {
  var lazyImages = [].slice.call(document.querySelectorAll("img.lazy"));

  let lazyImageObserver = new IntersectionObserver(function(entries, observer) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        let lazyImage = entry.target;
        lazyImage.src = lazyImage.dataset.src;
        lazyImage.srcset = lazyImage.dataset.srcset;
        lazyImage.classList.remove("lazy");
        lazyImageObserver.unobserve(lazyImage);
      }
    });
  });

  lazyImages.forEach(function(lazyImage) {
    lazyImageObserver.observe(lazyImage);
  });
});
