Game.Board = (function(g){
  "use strict";

  var tilesize = 16;
  var pz;
  var px;
  var py;

  var board = function(g, xt, yt) {
    this.game = g;
    this.xTiles = xt;
    this.yTiles = yt;
    this.i = Math.floor(xt/2);
    this.j = Math.floor(yt/2);
    this.dirty = true;
    this.tiles = [];
    // This will be replaced with tile loader.
    for (var i = 0; i < this.xTiles; i++) {
      this.tiles[i] = [];
      for (var j = 0; j < this.yTiles; j++) {
        this.tiles[i][j] = new Tile(tilesize, tilesize);
      }
    }
    this.tile = this.tiles[this.i][this.j];
  };

  board.prototype.render = function(ctx, cx, cy, curx, cury, zoom, dirty, mousedown, c, e) {
    var scale = tilesize*zoom;
    var x1 = parseInt(cx - this.xTiles*scale/2);
    var y1 = parseInt(cy - this.yTiles*scale/2);
    var x2 = x1 + this.xTiles*scale;
    var y2 = y1 + this.yTiles*scale;
    for (var i = 0; i < this.xTiles; i++) {
      for (var j = 0; j < this.yTiles; j++) {
        this.tiles[i][j].render(ctx, x1 + i * scale, y1 + j * scale, scale/tilesize, dirty);
      }
    }
    if (this.tile.inBounds(curx, cury) && !this.game.keyDown("alt") && !this.game.keyDown("tab")) {
      if (mousedown) {
        this.tile.setXY(curx, cury, c);
      } else {
        this.tile.cursor(ctx, curx, cury, c);
      }
    }
  };

  board.prototype.handleClick = function(e, cx, cy, curx, cury, zoom) {
    if (this.tile.inBounds(curx, cury)) {
      if (e.altKey) {
        this.game.setColor(this.tile.getXY(curx, cury));
      } else if (!this.game.keyDown("tab")) {
        this.tile.setXY(curx, cury, this.game.color());
      }
    }
  };

  var Tile = function(w, h){
    this.x1 = 0;
    this.y1 = 0;
    this.scale = 0;
    this.px = [];
    for (var i = 0; i < w; i++) {
      this.px[i] = [];
      for (var j = 0; j < h; j++) {
        this.px[i][j] = "#fff";
      }
    }
    this.dirty = true;
  };

  Tile.prototype.render = function(ctx, x1, y1, scale, dirty){
    if (dirty || this.dirty) {
      this.x1 = x1;
      this.y1 = y1;
      this.scale = scale;
      this.dirty = false;
      var prev = "";
      for (var i = 0; i < this.px.length; i++) {
        for (var j = 0; j < this.px[i].length; j++) {
          if (this.px[i][j] != prev) {
            ctx.fillStyle = this.px[i][j];
          }
          ctx.fillRect(x1 + i * scale, y1 + j * scale, scale, scale);
        }
      }
    }
  };

  Tile.prototype.get = function(i, j) {
    return this.px[i][j]
  };

  Tile.prototype.getXY = function(x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    return this.get(i, j);
  };

  Tile.prototype.set = function(i, j, c) {
    log(i, j, c);
    this.px[i][j] = c;
    this.dirty = true;
  };

  Tile.prototype.setXY = function(x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    this.set(i, j, c);
  };

  Tile.prototype.inBounds = function(x, y) {
    return this.x1 < x && x < this.x1 + tilesize * this.scale &&
      this.y1 < y && y < this.y1 + tilesize * this.scale
  };

  Tile.prototype.cursor = function(ctx, x, y, c) {
    this.drawPixel(ctx, x, y, c);
    this.dirty = true;
  };

  Tile.prototype.drawPixel = function(ctx, x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    ctx.fillStyle = c;
    ctx.fillRect(
      this.x1 + i * this.scale,
      this.y1 + j * this.scale,
      this.scale,
      this.scale,
    );
  }


  return board

})(Game);

