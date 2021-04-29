Game.Board = (function(g){
  "use strict";

  var tilesize = 16;

  var board = function(g, xt, yt) {
    this.game = g;
    this.xTiles = xt;
    this.yTiles = yt;
    this.i = 0;
    this.j = 0;
    this.dirty = true;
    this.scale = tilesize;
    this.tiles = [];
  };

  board.prototype.setData = function(bgctx) {
    for (var i = 0; i < this.xTiles; i++) {
      this.tiles[i] = [];
      for (var j = 0; j < this.yTiles; j++) {
        this.tiles[i][j] = new Game.Tile(bgctx.getImageData(i*tilesize, j*tilesize, tilesize, tilesize));
      }
    }
    this.tile = this.tiles[this.i][this.j];
  }

  board.prototype.render = function(ctx, cx, cy, curx, cury, zoom, dirty, mousedown, c, e) {
    if (this.tiles.length == 0) {
      return;
    }
    this.scale = tilesize*zoom;
    var x1 = parseInt(cx - (this.i+1)*this.scale + this.scale/2);
    var y1 = parseInt(cy - (this.j+1)*this.scale + this.scale/2);
    for (var i = 0; i < this.xTiles; i++) {
      for (var j = 0; j < this.yTiles; j++) {
        this.tiles[i][j].render(ctx, x1 + i * this.scale, y1 + j * this.scale, this.scale/tilesize, dirty || this.dirty);
      }
    }
    if (this.tile.active && this.tile.inBounds(curx, cury) && !this.game.keyDown("alt") && !this.game.keyDown("tab")) {
      this.tile.cursor(ctx, curx, cury, c);
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
      if (this.tile.active && !this.game.keyDown("tab")) {
        this.tile.setXY(curx, cury, this.game.color());
      }
    } else if (!this.tile || !this.tile.active && 0 <= i && i < this.xTiles && 0 <= j && j < this.yTiles) {
      this.i = i;
      this.j = j;
      this.tile = this.tiles[i][j];
      this.dirty = true;
    }
  };

  board.prototype.handleMouseMove = function(curx, cury, mousedown, c) {
    if (mousedown && this.tile && this.tile.active && this.tile.inBounds(curx, cury) &&
      !this.game.keyDown("alt") && !this.game.keyDown("tab")) {
      this.tile.setXY(curx, cury, c);
    }
  }

  board.prototype.moveTile = function(dx, dy) {
    this.dirty = true;
    if (this.tile.active) {
      this.tile.deactivate();
      // Commit edits
    }
    if ((dx < 0 && this.i < 1) || (dx > 0 && this.i >= this.xTiles - 1) ||
       ( dy < 0 && this.j < 1) || (dy > 0 && this.j >= this.yTiles - 1)) {
       return;
    }
    this.i += dx;
    this.j += dy;
    this.tile = this.tiles[this.i][this.j];
  }

  board.prototype.toggleActive = function() {
    this.dirty = true;
    this.tile.toggleActive();
  }

  board.prototype.isDirty = function() {
    return this.dirty;
  }

  return board

})(Game);

