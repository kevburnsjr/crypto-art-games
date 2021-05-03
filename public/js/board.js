Game.Board = (function(g){
  "use strict";

  var tilesize = 16;

  var board = function(g, src, palette, xt, yt) {
    this.game = g;
    this.xTiles = xt;
    this.yTiles = yt;
    this.prevx = -1;
    this.prevy = -1;
    this.i = 0;
    this.j = 0;
    this.dirty = true;
    this.scale = tilesize;
    this.palette = palette;
    this.tiles = [];
    this.edits = [];
    var icanvas = document.createElement('canvas');
    var ictx = icanvas.getContext("2d");
    var img = new Image();
    var self = this;
    img.onload = function() {
      icanvas.width = img.width;
      icanvas.height = img.height;
      ictx.drawImage(img, 0, 0);
      self.setData(ictx);
    };
    img.src = src;
    // temporary variable singleton to minimize garbage collection in render loop
    this.v = {};
  };

  board.prototype.setData = function(bgctx) {
    for (var i = 0; i < this.xTiles; i++) {
      this.tiles[i] = [];
      for (var j = 0; j < this.yTiles; j++) {
        this.tiles[i][j] = new Game.Tile(bgctx.getImageData(i*tilesize, j*tilesize, tilesize, tilesize), this.palette, i, j);
      }
    }
    this.tile = this.tiles[this.i][this.j];
  };

  board.prototype.render = function(ctx, cx, cy, curx, cury, zoom, dirty, mousedown, c, e) {
    if (this.tiles.length == 0) {
      return;
    }
    this.scale = tilesize*zoom;
    this.v.x1 = parseInt(cx - (this.i+1)*this.scale + this.scale/2);
    this.v.y1 = parseInt(cy - (this.j+1)*this.scale + this.scale/2);
    this.v.x2 = this.v.x1 + this.xTiles * this.scale;
    this.v.y2 = this.v.y1 + this.yTiles * this.scale;
    this.v.tileDirty = this.tile.dirty;
    for (var i = 0; i < this.xTiles; i++) {
      for (var j = 0; j < this.yTiles; j++) {
        this.v.tx = this.v.x1 + i * this.scale;
        this.v.ty = this.v.y1 + j * this.scale;
        if (this.v.tx + this.scale < 0 || this.v.ty + this.scale < 0 || this.v.tx > cx * 2 || this.v.ty > cy * 2) {
          // Don't render offscreen tiles
          continue;
        }
        this.tiles[i][j].render(ctx, this.v.tx, this.v.ty, this.scale/tilesize, dirty || this.dirty);
      }
    }
    if (this.tile.active && (this.tile.dirty || this.tile.inBounds(curx, cury)) && !this.game.isKeyDown("alt") && !this.game.isKeyDown("tab") && !this.game.isKeyDown("e")) {
      this.tile.cursor(ctx, curx, cury, c, this.v.tileDirty);
    } else if (this.tile.cursi > -1) {
      this.tile.clearCursor();
    }
    if (this.dirty || dirty) {
      this.tile.stroke(ctx);
    }
    this.dirty = false;
  };

  board.prototype.handleClick = function(e, cx, cy, curx, cury, zoom) {
    if (this.tiles.length == 0) {
      return;
    }
    var x1 = parseInt(cx - (this.i+1)*this.scale + this.scale/2);
    var y1 = parseInt(cy - (this.j+1)*this.scale + this.scale/2);
    var i = Math.floor((curx-x1) / this.scale);
    var j = Math.floor((cury-y1) / this.scale);
    if (e.altKey) {
      this.game.setColor(this.tiles[i][j].getXY(curx, cury));
      return
    }
    if (this.i == i && this.j == j) {
      if (this.tile.active && !this.game.isKeyDown("tab")) {
        if (this.game.isKeyDown("e")) {
          this.tile.clearXY(curx, cury);
        } else {
          this.tile.setXY(curx, cury, this.prevx, this.prevy, this.game.color());
        }
      }
    } else if (!this.tile || !this.tile.active && 0 <= i && i < this.xTiles && 0 <= j && j < this.yTiles) {
      this.i = i;
      this.j = j;
      this.tile = this.tiles[i][j];
      this.dirty = true;
    }
  };

  board.prototype.handleMouseMove = function(x, y, mousedown, c) {
    if (mousedown && this.tile && this.tile.active && this.tile.inBounds(x, y) &&
      !this.game.isKeyDown("alt") && !this.game.isKeyDown("tab")) {
      if (this.game.isKeyDown("e")) {
        this.tile.clearXY(x, y, this.prevx, this.prevy);
      } else {
        this.tile.setXY(x, y, this.prevx, this.prevy, c);
      }
    }
    this.prevx = x;
    this.prevy = y;
  };

  board.prototype.clearPath = function() {
    this.prevx = -1;
    this.prevy = -1;
  };

  board.prototype.moveTile = function(dx, dy) {
    var self = this;
    (this.tile.active ? this.tile.commit() : Promise.resolve()).then(function(tile) {
      if ((dx < 0 && self.i < 1) || (dx > 0 && self.i >= self.xTiles - 1) ||
         ( dy < 0 && self.j < 1) || (dy > 0 && self.j >= self.yTiles - 1)) {
         return;
      }
      self.i += dx;
      self.j += dy;
      self.tile = self.tiles[self.i][self.j];
      self.dirty = true;
    });
  };

  board.prototype.toggleActive = function() {
    var self = this;
    if (!this.tile.active) {
      return this.tile.lock().then(function(e){
        self.dirty = true;
        return true;
      });
    } else {
      return this.tile.commit().then(function(tile){
        self.edits.push(tile);
        self.dirty = true;
        return false;
      });
    }
  };

  board.prototype.togglePalette = function() {
    this.tile.dirty = true;
  };

  board.prototype.toggleEraser = function() {
    this.tile.dirty = true;
  };

  board.prototype.toggleDropper = function() {
    this.tile.dirty = true;
  };

  board.prototype.cancelActive = function() {
    var self = this;
    return (this.tile.active ? this.tile.rollback() : Promise.resolve()).then(function(){
      self.dirty = true;
    });
  };

  return board

})(Game);

