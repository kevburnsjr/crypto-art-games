Game.Tile = (function(g){
  "use strict";

  var maxScale = 32;
  var editLimit = 256;
  var artificialLatency = 250;

  var tile = function(imgData, palette, ti, tj, size){
    this.ti = ti;
    this.tj = tj;
    this.x1 = 0;
    this.y1 = 0;
    this.size = size;
    this.cursi = -1;
    this.cursj = -1;
    this.scale = 1;
    this.frames = [];
    this.px = [];
    this.maxScale = maxScale;
    this.canvas = document.createElement('canvas');
    // if ("OffscreenCanvas" in window) {
      // this.canvas = this.canvas.transferControlToOffscreen();
    // }
    this.ctx = this.canvas.getContext("2d");
    this.buffer = [];
    this.palette = palette;
    this.w = imgData ? imgData.width : this.size;
    this.h = imgData ? imgData.height : this.size;
    this.canvas.width = this.w * maxScale;
    this.canvas.height = this.h * maxScale;
    this.active = false;
    for (var i = 0; i < this.w; i++) {
      this.px[i] = [];
      this.buffer[i] = [];
      for (var j = 0; j < this.h; j++) {
        var n = j*this.w*4 + i*4;
        this.px[i][j] = imgData ? this.palette.getIdx(this.palette.nearestColor(imgData.data[n], imgData.data[n+1], imgData.data[n+2])) : null;
        this.buffer[i][j] = null;
      }
    }
    this.dirty = true;
    // temporary variable singleton to minimize garbage collection in render loop
    this.v = {};
  };

  tile.prototype.setBufferData = function(imgData){
    for (var i = 0; i < this.w; i++) {
      for (var j = 0; j < this.h; j++) {
        var n = j*this.w*4 + i*4;
        if (imgData.data[n+3] > 0) {
          this.set(i, j, "#" + this.palette.nearestColor(imgData.data[n], imgData.data[n+1], imgData.data[n+2]));
        }
      }
    }
    this.dirty = true;
  }

  tile.prototype.render = function(ctx, x1, y1, scale){
    if (scale != this.scale) {
      this.scale = scale;
    }
    if (this.dirty) {
      for (this.v.i = 0; this.v.i < this.w; this.v.i++) {
        for (this.v.j = 0; this.v.j < this.h; this.v.j++) {
          if (this.buffer[this.v.i][this.v.j] != null) {
            this.ctx.fillStyle = "#" + this.palette.colors[this.buffer[this.v.i][this.v.j]];
          } else {
            this.ctx.fillStyle = "#" + this.palette.colors[this.px[this.v.i][this.v.j]];
          }
          this.ctx.fillRect(this.v.i * this.maxScale, this.v.j * this.maxScale, this.maxScale, this.maxScale);
        }
      }
    }
    this.x1 = x1;
    this.y1 = y1;
    this.dirty = false;
    ctx.drawImage(this.canvas, x1, y1, this.w * scale, this.h * scale);
    window.renders++;
  };

  tile.prototype.lock = function() {
    var self = this;
    return Game.getSocket().lockTile(this).then(() => {
      self.active = true;
      self.dirty = true;
      self.cursi = -1;
      self.cursj = -1;
    });
  };

  tile.prototype.rollback = function() {
    var self = this;
    return Game.getSocket().unlockTile(this).then(() => {
      for (var i in self.buffer) {
        for (var j in self.buffer[i]) {
          self.buffer[i][j] = null;
        }
      }
      self.active = false;
      self.dirty = true;
    });
  };

  tile.prototype.commit = function() {
    if (!this.hasEdits()) {
      return this.rollback();
    }
    var f = new Game.Frame(this);
    this.active = false;
    this.clearBuffer();
    return Game.getSocket().sendFrame(f);
  };

  tile.prototype.clearBuffer = function() {
    for (var i in this.buffer) {
      for (var j in this.buffer[i]) {
        this.buffer[i][j] = null;
      }
    }
  };

  tile.prototype.renderFrameBuffer = function(f) {
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.width);
    var n = 0;
    var i = 0;
    for (let b of f.mask) {
      if (b == 1) {
        this.ctx.fillStyle = "#" + this.palette.colors[f.colors[i]];
        this.ctx.fillRect(Math.floor(n/this.size) * this.maxScale, n%this.size * this.maxScale, this.maxScale, this.maxScale);
        i++;
      }
      n++;
    }
  };

  tile.prototype.get = function(i, j) {
    var ci = this.buffer[i][j];
    if (ci == undefined) {
      ci = this.px[i][j];
    }
    return this.palette.colors[ci];
  };

  tile.prototype.getXY = function(x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    return this.get(i, j);
  };

  tile.prototype.set = function(i, j, c) {
    if (i < 0 || i >= this.w || j < 0 || j >= this.h) {
      return;
    }
    this.buffer[i][j] = this.palette.getIdx(c);
    this.dirty = true;
  };

  tile.prototype.hasEdits = function() {
    for (var i in this.buffer) {
      for (var j in this.buffer[i]) {
        if (this.buffer[i][j] != null && this.buffer[i][j] != this.px[i][j]) {
          return true;
        }
      }
    }
    return false;
  };

  tile.prototype.applyFrame = function(f, refresh) {
    if (!f) {
      return
    }
    var n = 0;
    var i = 0;
    var push = f.prev.length == 0 || refresh;
    for (let b of f.mask) {
      if (b == 1) {
        if (push) {
          f.prev.push(this.px[Math.floor(n/this.size)][n%this.size]);
        }
        this.px[Math.floor(n/this.size)][n%this.size] = f.colors[i];
        i++;
      }
      n++;
    }
    this.dirty = true;
  };

  tile.prototype.undoFrame = function(f) {
    if (!f) {
      return
    }
    var n = 0;
    var i = 0;
    for (let b of f.mask) {
      if (b == 1) {
        this.px[Math.floor(n/this.size)][n%this.size] = f.prev[i];
        i++;
      }
      n++;
    }
    this.dirty = true;
  };

  tile.prototype.setXY = function(x, y, prevx, prevy, c, brushSize) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    switch (brushSize) {
      case 1:
        this.set(i-1, j-1, c);
        this.set(i-1, j,   c);
        this.set(i-1, j+1, c);
        this.set(i,   j-1, c);
        this.set(i,   j+1, c);
        this.set(i+1, j-1, c);
        this.set(i+1, j,   c);
        this.set(i+1, j+1, c);
      case 0:
        this.set(i, j, c);
        break;
    }
    // Smoothing between mousemove samples
    if (prevx > -1) {
      // Delta between last pixel and current pixel
      var di = Math.floor((prevx-this.x1) / this.scale) - i;
      var dj = Math.floor((prevy-this.y1) / this.scale) - j;
      while (di != 0 || dj != 0) {
        if (-1 < i + di && i + di < (this.size-1)
         && -1 < j + dj && j + dj < (this.size-1)) {
          // Connect non-adjacent consecutive pixels
          switch (brushSize) {
            case 1:
              this.set(i-1 + di, j-1 + dj, c);
              this.set(i-1 + di, j + dj, c);
              this.set(i-1 + di, j+1 + dj, c);
              this.set(i + di, j-1 + dj, c);
              this.set(i + di, j+1 + dj, c);
              this.set(i+1 + di, j-1 + dj, c);
              this.set(i+1 + di, j + dj, c);
              this.set(i+1 + di, j+1 + dj, c);
            case 0:
              this.set(i + di, j + dj, c);
              break;
          }
        }
        di -= di ? di / Math.abs(di) : 0;
        dj -= dj ? dj / Math.abs(dj) : 0;
      }
    }
  };

  tile.prototype.clear = function(i, j) {
    if (i < 0 || i >= this.w || j < 0 || j >= this.h) {
      return;
    }
    if (this.buffer[i][j] == null) {
      return;
    }
    this.buffer[i][j] = null;
    this.dirty = true;
  };

  tile.prototype.clearXY = function(x, y, prevx, prevy, brushSize) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    switch (brushSize) {
      case 1:
        this.clear(i-1, j-1);
        this.clear(i-1, j);
        this.clear(i-1, j+1);
        this.clear(i,   j-1);
        this.clear(i,   j+1);
        this.clear(i+1, j-1);
        this.clear(i+1, j);
        this.clear(i+1, j+1);
      case 0:
        this.clear(i, j);
        break;
    }
    // Smoothing between mousemove samples
    if (prevx > -1) {
      // Delta between last pixel and current pixel
      var di = Math.floor((prevx-this.x1) / this.scale) - i;
      var dj = Math.floor((prevy-this.y1) / this.scale) - j;
      while (di != 0 || dj != 0) {
        if (-1 < i + di && i + di < (this.size-1)
         && -1 < j + dj && j + dj < (this.size)) {
          // Connect non-adjacent consecutive pixels
          switch (brushSize) {
            case 1:
              this.clear(i-1 + di, j-1 + dj);
              this.clear(i-1 + di, j   + dj);
              this.clear(i-1 + di, j+1 + dj);
              this.clear(i   + di, j-1 + dj);
              this.clear(i   + di, j+1 + dj);
              this.clear(i+1 + di, j-1 + dj);
              this.clear(i+1 + di, j   + dj);
              this.clear(i+1 + di, j+1 + dj);
            case 0:
              this.clear(i + di, j + dj);
              break;
          }
        }
        di -= di ? di / Math.abs(di) : 0;
        dj -= dj ? dj / Math.abs(dj) : 0;
      }
    }
  };

  tile.prototype.inBounds = function(x, y, brushSize) {
    const b = brushSize * this.scale;
    return (this.x1 - b) < x && x < (this.x1 + this.size * this.scale + b)
        && (this.y1 - b) < y && y < (this.y1 + this.size * this.scale + b);
  };

  tile.prototype.cursor = function(ctx, x, y, c, brushSize, dirty) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    if (!dirty && this.cursi == i && this.cursj == j) {
      return;
    }
    this.cursi = i;
    this.cursj = j;
    switch (brushSize) {
      case 1:
        this.drawPixel(ctx, i-1, j-1, c);
        this.drawPixel(ctx, i-1, j, c);
        this.drawPixel(ctx, i-1, j+1, c);
        this.drawPixel(ctx, i, j-1, c);
        this.drawPixel(ctx, i, j+1, c);
        this.drawPixel(ctx, i+1, j-1, c);
        this.drawPixel(ctx, i+1, j, c);
        this.drawPixel(ctx, i+1, j+1, c);
      case 0:
        this.drawPixel(ctx, i, j, c);
        break;
    }
    if (!dirty) {
      this.dirty = true;
    }
  };

  tile.prototype.clearCursor = function() {
    this.cursi = -10;
    this.cursj = -10;
    this.dirty = true;
  };

  tile.prototype.drawPixel = function(ctx, i, j, c) {
    if (i < 0 || i >= this.w || j < 0 || j >= this.h) {
      return;
    }
    ctx.fillStyle = c;
    ctx.fillRect(
      this.x1 + i * this.scale,
      this.y1 + j * this.scale,
      this.scale,
      this.scale,
    );
  };

  tile.prototype.stroke = function(ctx) {
    ctx.lineWidth = 1;
    if (this.active) {
      ctx.strokeStyle = "rgba(255,0,0,1)";
    } else {
      ctx.strokeStyle = "rgba(0,0,0,1)";
    }
    ctx.strokeRect(this.x1, this.y1, this.scale * this.px.length, this.scale * this.px.length);
  }

  tile.prototype.getID = function() {
    return this.ti * this.size + this.tj
  };

  return tile

})(Game);

