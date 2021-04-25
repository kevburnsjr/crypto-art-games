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
    } else if (!this.tile.active) {
      this.i = i;
      this.j = j;
      this.tile = this.tiles[this.i][this.j];
      this.dirty = true;
    }
  };

  board.prototype.handleMouseMove = function(curx, cury, mousedown, c) {
    if (mousedown && this.tile.active && this.tile.inBounds(curx, cury) &&
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
      this.cursorX = 0;
      this.cursorY = 0;
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

  Tile.prototype.deactivate = function() {
    this.active = false;
    this.dirty = true;
  };

  Tile.prototype.toggleActive = function() {
    this.active = !this.active;
    this.dirty = true;
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

  Tile.prototype.stroke = function(ctx) {
    ctx.lineWidth = .4;
    log(this.active);
    if (this.active) {
      ctx.strokeStyle = "rgba(255,0,0,1)";
    } else {
      ctx.strokeStyle = "rgba(0,0,0,0.5)";
    }
    ctx.strokeRect(this.x1, this.y1, this.scale * this.px.length, this.scale * this.px.length);
  }

  return board

})(Game);

