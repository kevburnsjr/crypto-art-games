var Game = (function(g){
  "use strict";

  var defaultZoom = 16
  var animationFrame;
  var deg_to_rad = Math.PI / 180.0;
  var color = "#000"
  var zoom = defaultZoom;
  var pzoom = defaultZoom;
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
    elem = canvas;
    ctx = elem.getContext('2d', {
      alpha: false,
      desynchronized: true
    });
    if(window.location.hash) {
      var parts = window.location.hash.substr(1).split(':');
      setColor(parts[0]);
      zoom = parseInt(parts[1]);
    } else {
      var el = paletteMenu.children[Math.floor(Math.random() * paletteMenu.children.length)];
      setColor(window.getComputedStyle(el).backgroundColor);
    }
    board = new Game.Board(Game, 16, 16);
    var icanvas = document.createElement('canvas');
    var ictx = icanvas.getContext("2d");
    var img = new Image();
    img.onload = function() {
      icanvas.width = img.width;
      icanvas.height = img.height;
      ictx.drawImage(img, 0, 0);
      board.setData(ictx);
    };
    img.src = "/palettes/autumn.gif";
    reset();
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
        dirty = true;
    }
    if (dirty || board.isDirty()) {
      ctx.clearRect(0, 0, w, h);
    }
    try {
      board.render(ctx, w/2, h/2, hoverX, hoverY, zoom, dirty, mousedown, color);
    } catch(e) {
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
  var worldnav = false;
  var brushState = false;
  document.addEventListener('mousedown', function(e){
    var t = e.target;
    if (t.parentElement.id == "logo") {
      e.preventDefault();
      document.getElementById("world-nav").classList.add("open");
      worldnav = true;
      return
    }
    if (t.id == "brush-state") {
      if (paletteMenu.style.display != "block") {
        paletteMenu.style.left = parseInt(t.offsetWidth*1.333);
        paletteMenu.style.removeProperty("top");
        paletteMenu.style.bottom = 0;
        paletteMenu.style.display = "block";
        brushState = true;
      }
      return
    }
    mousedown = true;
    clickpoint = [e.offsetX, e.offsetY];
    if (t.nodeName == "BUTTON" && t.parentElement.id == "palette") {
      e.preventDefault();
      e.stopPropagation();
      var color = window.getComputedStyle(t).backgroundColor.replaceAll(/%20/g,"");
      setColor(color);
      if (brushState) {
        paletteMenu.style.display = "none";
        brushState = false;
      }
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
    board.handleMouseMove(hoverX, hoverY, mousedown, color);
  });
  document.addEventListener('mouseup', function(e){
    if (worldnav) {
      document.getElementById("world-nav").classList.remove("open");
      worldnav = false;
      return
    }
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
    if (k == "e") {
      e.preventDefault();
      document.body.classList.add("erasing");
    }
    if (k == "w" || k == "arrowup") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(0, -1);
    }
    if (k == "a" || k == "arrowleft") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(-1, 0);
    }
    if (k == "s" || k == "arrowdown") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(0, 1);
    }
    if (k == "d" || k == "arrowright") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(1, 0);
    }
    if (k == "0" || k == "numpad0") {
      e.preventDefault();
      zoom = 2;
      setZoom();
    }
    if (k == "pageup") {
      e.preventDefault();
      // Navigate to next page
    }
    if (k == "pagedown") {
      e.preventDefault();
      // Navigate to previous page
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
    if (k == "escape") {
      e.preventDefault();
      board.cancelActive();
    }
  });
  document.addEventListener('keyup', function(e){
    var k = e.key.toLowerCase();
    keyDownMap[k] = false;
    if (k == "alt") {
      document.body.classList.remove("color-picking");
    }
    if (k == "e") {
      document.body.classList.remove("erasing");
    }
    if (k == "tab") {
      e.preventDefault();
      e.stopPropagation();
      if (paletteMenu.style.display != "none") {
        paletteMenu.style.display = "none";
      }
    }
  });
  document.addEventListener('keypress', function(e){
    if (e.key == " ") {
      e.preventDefault();
      board.toggleActive();
    }
  });
  document.addEventListener('wheel', function(e) {
    if (e.deltaY < 0) {
      zoom += Math.max(parseInt(zoom/2), 1);
    }
    if (e.deltaY > 0) {
      zoom -= Math.max(parseInt(zoom/2), 1);
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
    zoom = Math.max(1, Math.min(32, zoom));
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
    document.getElementById("brush-state").style.backgroundColor = color;
    // elem.style.   = color;
    sethash();
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
