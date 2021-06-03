Game.Board = (function(g){
  "use strict";

  var board = function(g, store, data, palette, callback) {
    this.id = data.id;
    this.active = data.active;
    this.finished = data.finished;
    this.tileSize = data.tsz;
    this.xTiles = data.w;
    this.yTiles = data.h;
    this.game = g;
    this.store = store;
    this.focused = false;
    this.prevx = -1;
    this.prevy = -1;
    this.undo = [];
    this.redo = [];
    this.i = 0;
    this.j = 0;
    this.dirty = true;
    this.scale = 1;
    this.palette = palette;
    this.tiles = [];
    this.frames = [];
    this.frameIdx = {};
    this.tile = null;
    this.edits = [];
    this.enabled = false;
    this.paused = false;
    this.timecode = 0;
    this.offset = 0;
    this.drawnOffset = 0;
    this.created = data.created;
    this.speed = 64;
    this.brush = 0;
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
    img.src = data.bg;
    // temporary variable collection to minimize garbage collection in render loop
    this.v = {renderTimecode: 0};
  };

  board.prototype.setData = function(bgctx) {
    for (var i = 0; i < this.xTiles; i++) {
      this.tiles[i] = [];
      for (var j = 0; j < this.yTiles; j++) {
        this.tiles[i][j] = new Game.Tile(bgctx.getImageData(i*this.tileSize, j*this.tileSize, this.tileSize, this.tileSize), this.palette, i, j, this.tileSize);
      }
    }
    this.tile = this.tiles[this.i][this.j];
  };

  board.prototype.render = function(ctx, uiCtx, cx, cy, curx, cury, zoom, dirty, uiDirty, mousedown, e) {
    if (!this.enabled) {
      return;
    }
    this.scale = this.tileSize*zoom;
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
          this.tiles[i][j].render(ctx, this.v.tx, this.v.ty, this.scale/this.tileSize);
        }
      }
    }
    this.v.diff = this.offset - this.drawnOffset;
    this.v.dir = this.v.diff == 0 ? 0 : this.v.diff / Math.abs(this.v.diff);
    if (this.v.dir == 1) {
      this.v.renderTimecode = Math.max(Math.min(this.offset, this.v.renderTimecode + (this.v.dir*this.speed)), 0);
    } else {
      this.v.renderTimecode = Math.max(Math.max(this.offset, this.v.renderTimecode + (this.v.dir*this.speed)), 0);
    }
    this.v.drawn = false;
    for (this.v.i = 0; this.drawnOffset < Math.floor(this.v.renderTimecode); this.v.i++) {
      this.applyFrame(this.frames[this.drawnOffset]);
      this.drawnOffset++;
      this.v.drawn = true;
    }
    for (this.v.i = 0; this.drawnOffset > Math.ceil(this.v.renderTimecode); this.v.i++) {
      this.undoFrame(this.frames[this.drawnOffset-1]);
      this.drawnOffset--;
      this.v.drawn = true;
    }
    if (this.drawnOffset - this.offset === 0 && this.v.drawn) {
      this.paused = this.offset != this.frames.length;
      g.nav().showRecent(this);
    }
    if (this.tile.active && (this.dirty || this.tile.dirty || this.tile.inBounds(curx, cury, this.brushSize())) && !this.game.isKeyDown("alt", "tab", "e")) {
      this.tile.cursor(ctx, curx, cury, this.palette.getColor(), this.brushSize(), this.v.tileDirty);
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
        this.palette.color = this.palette.getIdx(this.tiles[i][j].getXY(curx, cury));
        this.game.setColor();
      }
      return
    }
    if (this.focused && this.enabled && this.i == i && this.j == j) {
      if (this.paused) {
        this.offset = this.frames.length-1;
        this.paused = false;
      }
      if (this.tile.active && !this.game.isKeyDown("tab")) {
        if (this.game.isKeyDown("e")) {
          this.tile.clearXY(curx, cury, curx, cury, this.brushSize());
        } else {
          this.tile.setXY(curx, cury, this.prevx, this.prevy, this.palette.colors[this.palette.color], this.brushSize());
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

  board.prototype.brushSize = function() {
    const b = this.game.isKeyDown("shift") ? 1 : 0;
    if (this.brush != b) {
      this.brush = b;
      this.dirty = true;
      if (this.tile) {
        this.tile.dirty = true;
      }
    }
    return b;
  }

  board.prototype.handleMouseDown = function(cx, cy, curx, cury) {
    if (this.focused) {
      var x1 = parseInt(cx - (this.i+1)*this.scale + this.scale/2);
      var y1 = parseInt(cy - (this.j+1)*this.scale + this.scale/2);
    } else {
      var x1 = parseInt(cx - (this.xTiles * this.scale)/2);
      var y1 = parseInt(cy - (this.yTiles * this.scale)/2);
    }
    const i = Math.floor((curx-x1) / this.scale);
    const j = Math.floor((cury-y1) / this.scale);
    if(this.tile.active && (i != this.i || j != this.j) && !this.game.isKeyDown("shift", "tab", "alt", "e")) {
      this.commitActive();
    }
  };

  board.prototype.handleMouseMove = function(x, y, mousedown, c) {
    if (mousedown && this.tile && this.tile.active && this.tile.inBounds(x, y, this.brushSize()) &&
      !this.game.isKeyDown("alt", "tab")) {
      if (this.game.isKeyDown("e")) {
        this.tile.clearXY(x, y, this.prevx, this.prevy, this.brushSize());
      } else {
        this.tile.setXY(x, y, this.prevx, this.prevy, c, this.brushSize());
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
      const f = self.frames[self.frameIdx[timecode]];
      if (f === undefined) {
        rej();
        return;
      }
      game.socket().undoFrame(self, f).then((e) => {
        self.removeFrame(f);
      });
    });
  };

  board.prototype.commitActive = async function() {
    var self = this;
    return this.tile.active ? this.tile.commit().then((frame) => {
      self.redo = [];
      if (frame) {
        self.undo.push(frame.timecode);
      }
      Game.nav().flash("success", "Changes saved");
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
    return this.store.setItem("timecode", tc.toString(16).padStart(8, 0));
  };

  board.prototype.getOffset = function() {
    return this.offset;
  };

  board.prototype.setOffset = function(o) {
    this.offset = o;
  }

  board.prototype.saveFrame = async function(f) {
    if (this.enabled) {
      f.date = new Date((self.created + f.timestamp) * 1000);
      this.frameIdx[f.timecode] = this.frames.length;
      this.frames.push(f)
      this.tiles[f.ti][f.tj].frameIdx[f.timecode] = this.tiles[f.ti][f.tj].frames.length;
      this.tiles[f.ti][f.tj].frames.push(f);
      g.nav().updateScrubber(this.frames.length);
      if (!this.paused) {
        this.offset = this.frames.length-1;
      }
      this.timecode = f.timecode;
    }
    return this.store.setItem(f.timecode.toString(16).padStart(8, 0), f.toBytes());
  };

  board.prototype.removeFrame = async function(f) {
    await this.store.removeItem(f.timecode.toString(16).padStart(8, 0), f.toBytes());
    if (!(f.timecode in this.frameIdx)) {
      return;
    }
    if (this.enabled) {
      this.frames.splice(this.frameIdx[f.timecode],1);
      delete(this.frameIdx[f.timecode]);
      this.tiles[f.ti][f.tj].frames.splice(this.tiles[f.ti][f.tj].frameIdx[f.timecode],1);
      delete(this.tiles[f.ti][f.tj].frameIdx[f.timecode]);
      g.nav().updateScrubber(this.frames.length);
      this.offset = Math.min(this.frames.length, this.offset);
      if (!this.paused) {
        this.offset = this.frames.length-1;
      }
    }
    return;
  };

  board.prototype.applyFrame = function(f) {
    if (this.enabled && f && !f.deleted) {
      this.tiles[f.ti][f.tj].applyFrame(f);
    }
  };

  // TODO - Process user bans in batch
  board.prototype.applyUserBan = async function(ban) {
    var frameDate;
    var deleted = {};
    var i;
    var f;
    for(i = this.frames.length-1; i >= 0; i--) {
      f = this.frames[i];
      frameDate = (+f.date/1000).toFixed(0);
      console.log(f, f.date, frameDate, ban.since, ban.until, f.userid, ban.targetID);
      if (frameDate < ban.since) break;
      this.tiles[f.ti][f.tj].undoFrame(f);
      if (frameDate > ban.until) continue;
      if (f.userid != ban.targetID) continue;
      await this.removeFrame(f);
    }
    for(; i < this.frames.length; i++) {
      f = this.frames[i];
      f.resamplePrev(this.tiles[f.ti][f.tj]);
      this.applyFrame(f);
    }
    this.offset = Math.min(this.frames.length-1, this.offset);
  };

  board.prototype.undoFrame = function(f) {
    if (this.enabled && f) {
      this.tiles[f.ti][f.tj].undoFrame(f);
    }
  };

  board.prototype.scanFrames = async function(fn) {
    return this.store.iterate(function(v, k, i) {
      if (k.length == 8) {
        fn(parseInt(k, 16), v)
      }
    });
  };

  board.prototype.enable = function(timecode) {
    var self = this;
    if (this.enabled) {
      return;
    }
    return this.scanFrames(function(timecode, frameData) {
      const f = Game.Frame.fromBytes(frameData);
      f.date = new Date((self.created + f.timestamp) * 1000);
      self.offset = self.frames.length-1;
      self.frameIdx[f.timecode] = self.frames.length;
      self.frames.push(f);
      self.tiles[f.ti][f.tj].frameIdx[f.timecode] = self.tiles[f.ti][f.tj].frames.length;
      self.tiles[f.ti][f.tj].frames.push(f);
    }).then(() => {
      self.setTimecode(timecode);
    }).then(() => {
      g.nav().updateScrubber(self.frames.length);
      g.nav().showRecent(this);
      g.setColor();
      self.timecode = timecode;
      self.enabled = true;
    });
  };

  return board

})(Game);
