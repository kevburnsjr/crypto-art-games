Game.Palette = (function(g){
  "use strict";

  var palettesize = 16;
  var paletterowsize = 8;
  var defaultScale = 32;

  var palette = function(el, colors){
    this.el = el;
    this.canvas = el.querySelector('canvas');
    this.ctx = this.canvas.getContext("2d");
    this.scale = 0;
    this.colors = colors;
    this.colorIdx = {};
    this.w = paletterowsize;
    this.h = parseInt(colors.length / paletterowsize);
    this.active = false;
    this.nearest = nearestColor.from(colors);
    for(var i in colors) {
      this.colorIdx[colors[i]] = parseInt(i);
    }
  };

  palette.prototype.render = function(){
    this.canvas.width = this.w * this.scale;
    this.canvas.height = this.h * this.scale;
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    for (var i = 0; i < this.w; i++) {
      for (var j = 0; j < this.h; j++) {
        this.ctx.fillStyle = "#" + this.colors[j*8+i];
        this.ctx.fillRect(i * this.scale, j * this.scale, this.scale, this.scale);
      }
    }
  };

  palette.prototype.hide = function() {
    this.el.style.display = "none";
    this.active = false;
  };

  palette.prototype.show = function(x, y) {
    if (this.scale != defaultScale) {
      this.scale = defaultScale;
      this.render();
    }
    this.el.classList.remove("bottom");
    this.el.style.left = x - this.canvas.width / 2;
    this.el.style.top = y - this.canvas.height / 2;
    this.el.style.removeProperty("bottom");
    this.el.style.display = "block";
    this.active = true;
  };

  palette.prototype.showBottom = function() {
    if (this.scale != 2 * defaultScale) {
      this.scale = 2 * defaultScale;
      this.render();
    }
    this.el.classList.add("bottom");
    this.el.style.left = -5;
    this.el.style.removeProperty("top");
    this.el.style.bottom = -5;
    this.el.style.display = "block";
    this.active = true;
  };


  palette.prototype.update = function(colors) {
    this.colors = colors;
  };

  palette.prototype.getIdx = function(c) {
    if (c[0] == "#") {
      c = c.substr(1);
    } else if (c[0] == "r") {
      c = rgbToHex(c);
    }
    return this.colorIdx[c];
  };

  palette.prototype.getXY = function(x, y) {
    const b = this.canvas.getBoundingClientRect();
    var i = Math.floor((x-b.left) / this.scale);
    var j = Math.floor((y-b.top) / this.scale);
    return this.colors[j*paletterowsize+i];
  };

  palette.prototype.nearestColor = function(r, g, b) {
    return this.nearest({r: r, g: g, b: b});
  };

  var rgbToHex = function(str) {
    var res = /rgba?\((\d+)\,[^\d]*(\d+)\,[^\d]*(\d+).*/.exec(str);
    return ((1 << 24) + (parseInt(res[1]) << 16) + (parseInt(res[2]) << 8) + parseInt(res[3])).toString(16).slice(1);
  };

  return palette

})(Game);

