Game.Board = (function(g){
  "use strict";

  var tilesize = 16;

  var board = function(g, store, src, palette, xt, yt, callback) {
    this.game = g;
    this.store = store;
    this.focused = false;
    this.xTiles = xt;
    this.yTiles = yt;
    this.prevx = -1;
    this.prevy = -1;
    this.undo = [];
    this.redo = [];
    this.i = 0;
    this.j = 0;
    this.dirty = true;
    this.scale = tilesize;
    this.palette = palette;
    this.tiles = [];
    this.frames = [];
    this.tileFrameIndex = [];
    this.tile = null;
    this.edits = [];
    this.enabled = false;
    this.paused = false;
    this.timecode = 0;
    this.drawnTimecode = 0;
    this.timecheck = 0;
    this.speed = 4;
    var icanvas = document.createElement('canvas');
    var ictx = icanvas.getContext("2d");
    var img = new Image();
    var self = this;
    img.onload = function() {
      icanvas.width = img.width;
      icanvas.height = img.height;
      ictx.drawImage(img, 0, 0);
      self.setData(ictx);
      if (callback) {
        callback();
      }
    };
    img.src = src;
    // temporary variable collection to minimize garbage collection in render loop
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

  board.prototype.render = function(ctx, uiCtx, cx, cy, curx, cury, zoom, dirty, uiDirty, mousedown, c, e) {
    if (this.tiles.length == 0) {
      return;
    }
    this.scale = tilesize*zoom;
    if (this.focused) {
      this.v.x1 = parseInt(cx - (this.i+1)*this.scale + this.scale/2);
      this.v.y1 = parseInt(cy - (this.j+1)*this.scale + this.scale/2);
      this.v.x2 = this.v.x1 + this.xTiles * this.scale;
      this.v.y2 = this.v.y1 + this.yTiles * this.scale;
    } else {
      this.v.x1 = parseInt(cx - (this.xTiles * this.scale)/2);
      this.v.y1 = parseInt(cy - (this.yTiles * this.scale)/2);
      this.v.x2 = this.v.x1 + this.xTiles * this.scale;
      this.v.y2 = this.v.y1 + this.yTiles * this.scale;
    }
    this.v.tileDirty = this.tile.dirty;
    for (var i = 0; i < this.xTiles; i++) {
      for (var j = 0; j < this.yTiles; j++) {
        this.v.tx = this.v.x1 + i * this.scale;
        this.v.ty = this.v.y1 + j * this.scale;
        if (this.v.tx + this.scale < 0 || this.v.ty + this.scale < 0 || this.v.tx > cx * 2 || this.v.ty > cy * 2) {
          // Don't render offscreen tiles
          continue;
        }
        if (dirty || this.dirty || this.tiles[i][j].dirty) {
          this.tiles[i][j].render(ctx, this.v.tx, this.v.ty, this.scale/tilesize);
        }
      }
    }
    var drawn = this.drawnTimecode - this.timecode != 0;
    if (this.enabled) {
      for (this.v.i = 0; this.drawnTimecode < this.timecode && this.v.i < this.speed; this.v.i++) {
        this.applyFrame(this.frames[this.drawnTimecode]);
        this.drawnTimecode++;
      }
      for (this.v.i = 0; this.drawnTimecode > this.timecode && this.v.i < this.speed; this.v.i++) {
        this.undoFrame(this.frames[this.drawnTimecode-1]);
        this.drawnTimecode--;
      }
    }
    if (this.drawnTimecode - this.timecode === 0 && drawn) {
      g.nav().showRecent(this);
      this.paused = this.timecode != this.frames.length;
    }
    if (this.tile.active && (this.tile.dirty || this.tile.inBounds(curx, cury)) && !this.game.isKeyDown("alt") && !this.game.isKeyDown("tab") && !this.game.isKeyDown("e")) {
      this.tile.cursor(ctx, curx, cury, c, this.v.tileDirty);
    } else if (this.tile.cursi > -1) {
      this.tile.clearCursor();
    }
    if (this.uiDirty || uiDirty) {
      uiCtx.clearRect(0, 0, uiCtx.canvas.width, uiCtx.canvas.height);
      if (this.focused) {
        this.tile.stroke(uiCtx);
      }
    }
    this.uiDirty = false;
    this.dirty = false;
  };

  board.prototype.handleClick = function(e, cx, cy, curx, cury, zoom) {
    if (this.tiles.length == 0) {
      return;
    }
    if (this.focused) {
      var x1 = parseInt(cx - (this.i+1)*this.scale + this.scale/2);
      var y1 = parseInt(cy - (this.j+1)*this.scale + this.scale/2);
    } else {
      var x1 = parseInt(cx - (this.xTiles * this.scale)/2);
      var y1 = parseInt(cy - (this.yTiles * this.scale)/2);
    }
    var i = Math.floor((curx-x1) / this.scale);
    var j = Math.floor((cury-y1) / this.scale);
    if (e.altKey) {
      if (i >= 0 && i < this.xTiles && j >= 0 && j < this.yTiles) {
        this.game.setColor(this.tiles[i][j].getXY(curx, cury));
      }
      return
    }
    if (e.button == 2) {
      e.preventDefault();
      if (this.tile.active) {
        if (this.i == i && this.j == j) {
          var self = this;
          this.tile.commit().then(function(f){
            self.dirty = true;
          }).catch((e) => {
            self.cancelActive();
          });
        } else {
          this.cancelActive();
        }
      }
      return
    }
    if (this.focused && this.enabled && this.i == i && this.j == j) {
      if (this.paused) {
        this.timecode = this.frames.length;
        this.paused = false;
      }
      if (this.tile.active && !this.game.isKeyDown("tab")) {
        if (this.game.isKeyDown("e")) {
          this.tile.clearXY(curx, cury);
        } else {
          this.tile.setXY(curx, cury, this.prevx, this.prevy, this.game.color());
        }
      } else if(!this.tile.active) {
        this.toggleActive();
      }
    } else if (!this.tile || (!this.tile.active && 0 <= i && i < this.xTiles && 0 <= j && j < this.yTiles)) {
      this.setFocus(i, j);
    } else {
      this.cancelFocus();
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

  board.prototype.setTile = function(n) {
    if (n <= 255 && n >= 0) {
      this.setFocus(Math.floor(n/this.xTiles), n % this.xTiles);
    }
  };

  board.prototype.getTileID = function() {
    return this.i*16 + this.j;
  };

  board.prototype.moveTile = function(dx, dy) {
    var self = this;
    return this.commitActive().then(function(tile) {
      self.setFocus(self.i + dx, self.j + dy);
    });
  };

  board.prototype.toggleActive = async function() {
    if (!this.enabled) {
      return Promise.resolve();
    }
    var self = this;
    if (this.focused && !this.tile.active) {
      return this.tile.lock().then(function(e){
        document.body.classList.add("editing");
        self.uiDirty = true;
        return true;
      }).catch((e) => {
        return false;
      });
    } else if (this.tile.active) {
      return this.commitActive().then(function(){
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

  board.prototype.undo = async function() {
    var self = this;
    return new Promise((res, rej) => {
      if (this.tile.active) {
        rej();
        return;
      }
      const timecode = self.undo.shift();
      if (timecode === undefined) {
        rej();
        return;
      }
      const f = self.frames[timecode];
      if (f === undefined) {
        rej();
        return;
      }
      game.socket().undoFrame(self, f).then((e) => {
        f.deleted = true;
        g.nav().updateScrubber(f.timecode);
      });
    });
  };

  board.prototype.commitActive = async function() {
    var self = this;
    return this.tile.active ? this.tile.commit().then((frame) => {
      self.redo = [];
      self.undo.push(frame.timecode);
      self.uiDirty = true;
    }) : Promise.resolve();
  };

  board.prototype.cancelActive = async function() {
    var self = this;
    return (this.tile.active ? this.tile.rollback() : Promise.resolve()).then(function(){
      self.uiDirty = true;
    });
  };

  board.prototype.setFocus = function(i, j) {
    if (i < 0 || i > this.xTiles - 1 || j < 0 || j > this.yTiles - 1) {
       return;
    }
    this.i = i;
    this.j = j;
    this.tile = this.tiles[i][j];
    this.dirty = true;
    this.uiDirty = true;
    this.focused = true;
    if (this.game.nav() != undefined) {
      this.game.nav().showRecent(this);
    }
  };

  board.prototype.cancelFocus = function() {
    if (this.tile && this.tile.active) {
      return;
    }
    this.focused = false;
    this.dirty = true;
    this.uiDirty = true;
    if (this.game.nav() != undefined) {
      this.game.nav().showRecent(this);
    }
  };

  board.prototype.getTimecode = async function(tc) {
    return this.store.getItem("timecode").then(t => t ? parseInt(t, 16) : 0);
  };

  board.prototype.setTimecode = async function(tc) {
    return this.store.setItem("timecode", tc.toString(16).padStart(4, 0));
  };

  board.prototype.setUserIdx = async function(userIdx) {
    return this.store.setItem("userIdx", userIdx.toString(16).padStart(4, 0));
  };

  board.prototype.saveFrame = async function(f) {
    if (this.enabled) {
      if (f.timecheck > 0) {
        this.timecheck = f.timecheck*1000;
      }
      f.date = new Date();
      this.frames[f.timecode] = f;
      this.tiles[f.ti][f.tj].frames.push(f);
      g.nav().updateScrubber(f.timecode+1);
      if (!this.paused) {
        this.timecode = f.timecode+1;
      }
    }
    return this.store.setItem(f.timecode.toString(16).padStart(4, 0), f.toBytes());
  };

  board.prototype.scanFrames = async function(fn) {
    return this.store.iterate(function(v, k, i) {
      if (k.length == 4) {
        fn(parseInt(k, 16), v)
      }
    });
  };

  board.prototype.applyFrame = function(f) {
    if (this.enabled && f) {
      this.tiles[f.ti][f.tj].applyFrame(f);
    }
  };

  board.prototype.undoFrame = function(f) {
    if (this.enabled && f) {
      this.tiles[f.ti][f.tj].undoFrame(f);
    }
  };

  board.prototype.enable = function(timecode, userIdx, bucket) {
    var self = this;
    if (this.enabled) {
      return;
    }
    return this.scanFrames(function(timecode, frameData) {
      const f = Game.Frame.fromBytes(frameData);
      if (f.timecheck > 0) {
        self.timecheck = f.timecheck*1000;
      }
      f.date = new Date(self.timecheck + f.timestamp * 60 * 1000);
      self.frames[timecode] = f;
      self.tiles[f.ti][f.tj].frames.push(f);
    }).then(() => {
      self.setTimecode(timecode);
    }).then(() => {
      self.setUserIdx(userIdx);
    }).then(() => {
      g.nav().updateScrubber(timecode);
      self.timecode = timecode;
      self.enabled = true;
    });
  };

  return board

})(Game);

