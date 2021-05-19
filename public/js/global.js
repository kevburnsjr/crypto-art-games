(function() {
  "use strict";

  document.addEventListener("DOMContentLoaded", function(e) {
    Game.start(
      document.getElementById("bgcanvas"),
      document.getElementById("palette")
    );
  });
  document.querySelectorAll("#left nav a").forEach(el => {
    var id = el.dataset.toggle;
    el.addEventListener("click", (e) => {
      e.preventDefault();
      e.target.classList.toggle('active');
      document.getElementById(id).classList.toggle('active');
      localforage.setItem("ui-"+id, e.target.classList.contains('active'));
    });
    localforage.getItem("ui-"+id).then(active => {
      if (active == null) {
        localforage.setItem("ui-"+id, true);
        active = true
      }
      if (active) {
        el.click();
      }
    });
  });

  var lto;
  window.debug = function(){
    console.log("Renders:", window.renders);
    window.renders = 0;
    lto = setTimeout(window.debug, 1000);
  };

})();

// convert rgb to hsl (ie. dark/light cursor selection)
(function() {
  window.rgbToHsl = function(r, g, b) {
    r /= 255, g /= 255, b /= 255;
    var max = Math.max(r, g, b), min = Math.min(r, g, b);
    var h, s, l = (max + min) / 2;
    if (max == min) {
      h = s = 0; // achromatic
    } else {
      var d = max - min;
      s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
      switch (max) {
        case r: h = (g - b) / d + (g < b ? 6 : 0); break;
        case g: h = (b - r) / d + 2; break;
        case b: h = (r - g) / d + 4; break;
      }
      h /= 6;
    }

    return [ h, s, l ];
  }
})();
