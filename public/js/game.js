var Game = (function(g){
  "use strict";

  var animationFrame;
  var deg_to_rad = Math.PI / 180.0;
  var color = "#fff"
  var zoom = 16;
  var pzoom = 16;
  var ctx, elem;
  var last = Date.now();
  var fps_tick = Date.now();
  var pause = false;
  var animate = true;
  var w;
  var h;
  var board;
  var hoverX;
  var hoverY;
  var bgtimeout = null;
  var paletteMenu;

  var start = function(canvas, p) {
    paletteMenu = p;
    console.log("start");
    elem = canvas;
    ctx = elem.getContext('2d');
    if(window.location.hash) {
      var parts = window.location.hash.substr(1).split(':');
      setColor(parts[0]);
      zoom = parseInt(parts[1]);
    }
    board = new Game.Board(Game, 16, 9);
    reset();
    // Load board(s)
  };

  var reset = function() {
    setZoom();
    w = window.innerWidth;
    h = window.innerHeight;
    draw();
  };

  var draw = function() {
    animationFrame = window.requestAnimationFrame(draw);
    var dirty = false;
    if (elem.width != w || elem.height != h || zoom != pzoom) {
        elem.width = w;
        elem.height = h;
        pzoom = zoom;
        ctx.clearRect(0, 0, w, h);
        dirty = true;
    }
    try {
      board.render(ctx, w/2, h/2, hoverX, hoverY, zoom, dirty, mousedown, color);
    } catch(e) {
      console.log(e);
      window.cancelAnimationFrame(animationFrame);
      bgtimeout = setTimeout(function(){
        window.cancelAnimationFrame(animationFrame);
        draw();
      }, 1000);
    }

    var now = Date.now();
    if(fps_tick + 1000 < now) {
      // $('#fps').text(Math.round(1000/(now - last)) + " fps");
      fps_tick = now;
    }
    last = now;
  };

  // ----------------- Input Functions -------------------

  var clickpoint = [];
  var mousedown = false;
  document.addEventListener('mousedown', function(e){
    mousedown = true;
    clickpoint = [e.offsetX, e.offsetY];
    var t = e.target;
    if (t.nodeName == "BUTTON" && t.parentElement.id == "palette") {
      e.preventDefault();
      e.stopPropagation();
      log(e)
      var color = window.getComputedStyle(t).backgroundColor.replaceAll(/%20/g,"");
      Game.setColor(color);
    } else {
      board.handleClick(e, w/2, h/2, hoverX, hoverY, zoom)
    }
    if (e.target.nodeName != "CANVAS") {
      return;
    }
  });
  document.addEventListener('mousemove', function(e){
    hoverX = Math.round(e.pageX);
    hoverY = Math.round(e.pageY);
  });
  document.addEventListener('mouseup', function(e){
    mousedown = false;
    clickpoint = [];
    animate = true;
  });
  var keyDownMap = {};
  var keyDown = function(k) {
    return keyDownMap[k];
  }
  document.addEventListener('keydown', function(e){
    var k = e.key.toLowerCase();
    keyDownMap[k] = true;
    if (k == "alt") {
      e.preventDefault();
      document.body.classList.add("color-picking");
    }
    if (k == "tab") {
      e.preventDefault();
      e.stopPropagation();
      if (paletteMenu.style.display != "block") {
        paletteMenu.style.left = hoverX;
        paletteMenu.style.top = hoverY;
        paletteMenu.style.display = "block";
      }
    }
  });
  document.addEventListener('keyup', function(e){
    var k = e.key.toLowerCase();
    keyDownMap[k] = false;
    if (k == "alt") {
      document.body.classList.remove("color-picking");
    }
    if (k == "tab") {
      e.preventDefault();
      e.stopPropagation();
      if (paletteMenu.style.display != "none") {
        paletteMenu.style.display = "none";
      }
    }
  });
  document.addEventListener('wheel', function(e) {
    if (e.deltaY < 0) {
      zoom += 2;
    }
    if (e.deltaY > 0) {
      zoom -= 2;
    }
    setZoom();
    sethash();
  });

  // ----------------- View Functions -------------------

  window.addEventListener('resize', function(e){
    clearTimeout(bgtimeout);
    bgtimeout = setTimeout(function(){
      w = window.innerWidth;
      h = window.innerHeight;
    }, 300);
  });

  // ----------------- State Functions -------------------

  var sethash = function() {
    window.location.hash = [color, zoom].join(':');
  };

  window.onhashchange = function() {
    reset();
  };

  var setZoom = function() {
    zoom = Math.max(1, Math.min(64, zoom));
  };

  var setColor = function(c) {
    color = c.replaceAll(/%20/g,"");
    var rgb = [...color.matchAll(/\d+/g)];
    var hsl = rgbToHsl(rgb[0], rgb[1], rgb[2]);
    if (hsl[2] > 0.5) {
      document.body.classList.add('bg-light');
    } else {
      document.body.classList.remove('bg-light');
    }
    elem.style.backgroundColor = color;
  };

  return {
    start: start,
    color: function(){
      return color.replaceAll(/%20/g, "");
    },
    mousedown: function(){
      return mousedown;
    },
    setColor: setColor,
    keyDown: keyDown
  };

})(Game || {});
