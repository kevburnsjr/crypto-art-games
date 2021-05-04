var Game = (function(g){
  "use strict";

  var defaultZoom = 16;
  var animationFrame;
  var deg_to_rad = Math.PI / 180.0;
  var bgcolor = "#666";
  var color = "#000";
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
  var palette;
  var hoverX;
  var hoverY;
  var bgtimeout = null;
  var autumn = [
    "ec6f1c", "b4522e", "7a3030", "f6ae3c", "fbdb7a", "eafba3", "e3f6d5", "9ce77f",
    "49d866", "408761", "2d4647", "345452", "3a878b", "3da4db", "95c5f2", "cacff9"
  ];
  var last;
  var socket;

  var start = function(canvasElem, paletteElem) {
    elem = canvasElem;
    ctx = elem.getContext('2d', {
      alpha: false,
      desynchronized: true
    });
    if(window.location.hash) {
      var parts = window.location.hash.substr(1).split(':');
      setColor(parts[0]);
      zoom = parseInt(parts[1]);
    } else {
      setColor(autumn[Math.floor(Math.random() * autumn.length)]);
    }
    palette = new Game.Palette(paletteElem, autumn);
    board = new Game.Board(Game, "/palettes/autumn.gif", palette, 16, 16);
    reset();
    document.addEventListener('mousemove', mousemove);
    document.addEventListener('mousedown', mousedown);
    document.addEventListener('mouseup', mouseup);
    document.addEventListener('keydown', keydown);
    document.addEventListener('keyup', keyup);
    document.addEventListener('keypress', keypress);
    document.addEventListener('wheel', wheel);
    document.addEventListener('visibilitychange', visibilitychange);
    window.addEventListener('resize', resize);
    // initiate websocket
    socket = new Game.socket({
      url: function() {
        return "/socket?ts="+(last || 0);
      }
    });
    socket.on('message', function(e) {
      console.log(e);
    });
    socket.start();
  };

  var reset = function() {
    setZoom();
    w = window.innerWidth;
    h = window.innerHeight;
    draw();
  };

  var draw = function() {
    window.cancelAnimationFrame(animationFrame);
    animationFrame = window.requestAnimationFrame(draw);
    var dirty = false;
    if (elem.width != w || elem.height != h || zoom != pzoom) {
        elem.width = w;
        elem.height = h;
        pzoom = zoom;
        dirty = true;
    }
    if (dirty || board.dirty) {
      ctx.fillStyle = bgcolor;
      ctx.fillRect(0, 0, w, h);
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
  var isMousedown = false;
  var worldnav = false;
  var brushState = false;
  var keyDownMap = {};
  var isKeyDown = function(k) {
    return keyDownMap[k];
  }

  // mousedown
  var mousedown = function(e){
    var t = e.target;
    if (t.parentElement.id == "logo") {
      e.preventDefault();
      document.getElementById("world-nav").classList.add("open");
      worldnav = true;
      return
    }
    if (t.id == "brush-state") {
      if (!palette.active) {
        palette.showBottom();
        brushState = true;
      }
      return
    }
    isMousedown = true;
    clickpoint = [e.offsetX, e.offsetY];
    if (t.id == "palette") {
      e.preventDefault();
      e.stopPropagation();
      setColor(palette.getXY(e.pageX, e.pageY));
      if (brushState) {
        palette.hide();
        brushState = false;
      }
    } else {
      board.handleClick(e, w/2, h/2, hoverX, hoverY, zoom)
    }
    if (e.target.nodeName != "CANVAS") {
      return;
    }
  };

  // mousemove
  var mousemove = function(e){
    hoverX = Math.round(e.pageX);
    hoverY = Math.round(e.pageY);
    board.handleMouseMove(hoverX, hoverY, isMousedown, color);
  };

  // mouseup
  var mouseup = function(e){
    if (worldnav) {
      document.getElementById("world-nav").classList.remove("open");
      worldnav = false;
      return
    }
    isMousedown = false;
    board.clearPath();
    clickpoint = [];
  };

  // keydown
  var keydown = function(e){
    var k = e.key.toLowerCase();
    keyDownMap[k] = true;
    if (k == "alt") {
      e.preventDefault();
      if (!document.body.classList.contains("color-picking")) {
        document.body.classList.add("color-picking");
        board.toggleDropper();
      }
    }
    if (k == "e") {
      e.preventDefault();
      if (!document.body.classList.contains("erasing")) {
        document.body.classList.add("erasing");
        board.toggleEraser();
      }
    }
    if (k == "tab") {
      e.preventDefault();
      e.stopPropagation();
      if (!palette.active) {
        palette.show(hoverX, hoverY);
        board.togglePalette();
      }
    }
    if (k == "w" || k == "arrowup") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(0, -1);
      document.body.classList.remove("editing");
    }
    if (k == "a" || k == "arrowleft") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(-1, 0);
      document.body.classList.remove("editing");
    }
    if (k == "s" || k == "arrowdown") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(0, 1);
      document.body.classList.remove("editing");
    }
    if (k == "d" || k == "arrowright") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(1, 0);
      document.body.classList.remove("editing");
    }
    if (k == "0" || k == "numpad0") {
      e.preventDefault();
      zoom = 2;
      setZoom();
    }
    if (k == "pageup") {
      e.preventDefault();
      // Navigate to next board
    }
    if (k == "pagedown") {
      e.preventDefault();
      // Navigate to previous board
    }
    if (k == "escape") {
      e.preventDefault();
      board.cancelActive().then(function() {
        document.body.classList.remove("editing");
      });
      if (brushState) {
        palette.hide();
        brushState = false;
      }
    }
  };

  // keyup
  var keyup = function(e){
    var k = e.key.toLowerCase();
    keyDownMap[k] = false;
    if (k == "alt") {
      document.body.classList.remove("color-picking");
      board.toggleDropper();
    }
    if (k == "e") {
      document.body.classList.remove("erasing");
      board.toggleEraser();
    }
    if (k == "tab") {
      e.preventDefault();
      e.stopPropagation();
      palette.hide();
      board.togglePalette();
    }
  };

  // keypress
  var keypress = function(e){
    if (e.key == " ") {
      e.preventDefault();
      board.toggleActive().then(function(active){
        if (active) {
          document.body.classList.add("editing");
        } else {
          document.body.classList.remove("editing");
        }
      });
    }
  };

  // wheel
  var wheel = function(e) {
    if (e.deltaY < 0) {
      if (zoom < 6) {
        zoom += 1;
      } else if (zoom < 12) {
        zoom += 2;
      } else {
        zoom += 4;
      }
    }
    if (e.deltaY > 0) {
      if (zoom <= 6) {
        zoom -= 1;
      } else if (zoom <= 12) {
        zoom -= 2;
      } else {
        zoom -= 4;
      }
    }
    setZoom();
    sethash();
  };

  // ----------------- View Functions -------------------

  var resize = function(e){
    clearTimeout(bgtimeout);
    bgtimeout = setTimeout(function(){
      w = window.innerWidth;
      h = window.innerHeight;
    }, 100);
  };
  var visibilitychange = function(e){
    document.getElementById("world-nav").classList.remove("open");
    document.body.classList.remove("color-picking");
    document.body.classList.remove("erasing");
    document.body.classList.remove("editing");
    board.cancelActive();
  };

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
    color = c.replace(/%20/g,"");
    var rgb = [...color.matchAll(/\d+/g)];
    if (c.length == 6) {
      color = "#" + color;
      rgb = [parseInt(c.substr(0,2), 16), parseInt(c.substr(2,2), 16), parseInt(c.substr(4,2), 16)];
    }
    var hsl = rgbToHsl(rgb[0], rgb[1], rgb[2]);
    if (hsl[2] > 0.5) {
      document.body.classList.add('bg-light');
    } else {
      document.body.classList.remove('bg-light');
    }
    document.getElementById("brush-state").style.backgroundColor = color;
    sethash();
  };

  var log = function() {
    if ("console" in window) {
      console.log(...args);
    }
  };

  var getSocket = function() {
    return socket;
  };

  return {
    start: start,
    color: function(){
      return color.replace(/%20/g, "");
    },
    mousedown: function(){
      return mousedown;
    },
    setColor: setColor,
    isKeyDown: isKeyDown,
    online: false,
    log: log,
    getSocket: getSocket
  };

})({});
