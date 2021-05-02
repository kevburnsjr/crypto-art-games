Game.Tile = (function(g){
  "use strict";

  var tilesize = 16;
  var editLimit = 256;
  var artificialLatency = 250;

  var tile = function(imgData, palette){
    this.x1 = 0;
    this.y1 = 0;
    this.cursi = -1;
    this.cursj = -1;
    this.scale = 1;
    this.px = [];
    this.canvas = document.createElement('canvas');
    this.ctx = this.canvas.getContext("2d");
    this.buffer = [];
    this.bufferCount = 0;
    this.palette = palette;
    this.w = imgData.width;
    this.h = imgData.height;
    this.active = false;
    for (var i = 0; i < this.w; i++) {
      this.px[i] = [];
      this.buffer[i] = [];
      for (var j = 0; j < this.h; j++) {
        var n = j*this.w*4 + i*4;
        this.px[i][j] = "rgb("+imgData.data[n]+","+imgData.data[n+1]+","+imgData.data[n+2]+")";
        this.buffer[i][j] = "";
      }
    }
    this.dirty = true;
  };

  tile.prototype.render = function(ctx, x1, y1, scale, dirty){
    if (scale != this.scale) {
      this.scale = scale;
      this.canvas.width = scale * this.w;
      this.canvas.height = scale * this.h;
      this.dirty = true;
      console.log(scale, this.scale);
    }
    if (this.dirty) {
      for (var i = 0; i < this.w; i++) {
        for (var j = 0; j < this.h; j++) {
          if (this.buffer[i][j] != "") {
            this.ctx.fillStyle = this.buffer[i][j];
            this.ctx.fillRect(i * scale, j * scale, scale, scale);
            continue
          }
          if (this.ctx.fillStyle != this.px[i][j]) {
            this.ctx.fillStyle = this.px[i][j];
          }
          this.ctx.fillRect(i * scale, j * scale, scale, scale);
        }
      }
    }
    if (dirty || this.dirty) {
      this.x1 = x1;
      this.y1 = y1;
      this.dirty = false;
      ctx.drawImage(this.canvas, x1, y1)
      window.renders++;
    }
  };

  tile.prototype.lock = function() {
    var self = this;
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        self.active = true;
        self.dirty = true;
        self.cursi = -1;
        self.cursj = -1;
        resolve();
      }, artificialLatency);
    });
  };

  tile.prototype.rollback = function() {
    var self = this;
    this.active = false;
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        for (var i in self.buffer) {
          for (var j in self.buffer[i]) {
            if (self.buffer[i][j] != "") {
              self.buffer[i][j] = "";
            }
          }
        }
        self.active = false;
        self.dirty = true;
        self.bufferCount = 0;
        resolve();
      }, artificialLatency);
    });
  };

  tile.prototype.commit = function() {
    var self = this;
    var f = new Game.Frame(this.palette, this.buffer);
    this.active = false;
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        for (var i in self.buffer) {
          for (var j in self.buffer[i]) {
            if (self.buffer[i][j] != "") {
              self.px[i][j] = self.buffer[i][j];
              self.buffer[i][j] = "";
            }
          }
        }
        var timecode = 12345;
        self.dirty = true;
        self.bufferCount = 0;
        resolve(timecode);
      }, artificialLatency);
    });
  };

  tile.prototype.get = function(i, j) {
    return this.buffer[i][j] || this.px[i][j]
  };

  tile.prototype.getXY = function(x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    return this.get(i, j);
  };

  tile.prototype.set = function(i, j, c) {
    if (this.px[i][j] == c) {
      this.clear(i, j);
      return;
    }
    if (this.buffer[i][j] == "") {
      if (this.bufferCount > editLimit) {
        return;
      }
      this.bufferCount++;
    }
    this.buffer[i][j] = c;
    this.dirty = true;
  };

  tile.prototype.setXY = function(x, y, prevx, prevy, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    this.set(i, j, c);
    // Smoothing between mousemove samples
    if (prevx > -1) {
      // Delta between last pixel and current pixel
      var di = Math.floor((prevx-this.x1) / this.scale) - i;
      var dj = Math.floor((prevy-this.y1) / this.scale) - j;
      while (di != 0 || dj != 0) {
        if (-1 < i + di && i + di < (tilesize-1)
         && -1 < j + dj && j + dj < (tilesize-1)) {
          // Connect non-adjacent consecutive pixels
          this.set(i + di, j + dj, c);
        }
        di -= di ? di / Math.abs(di) : 0;
        dj -= dj ? dj / Math.abs(dj) : 0;
      }
    }
  };

  tile.prototype.clear = function(i, j) {
    if (this.buffer[i][j] == "") {
      return;
    }
    this.buffer[i][j] = "";
    this.bufferCount--;
    this.dirty = true;
  };

  tile.prototype.clearXY = function(x, y, prevx, prevy) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    this.clear(i, j);
    // Smoothing between mousemove samples
    if (prevx > -1) {
      // Delta between last pixel and current pixel
      var di = Math.floor((prevx-this.x1) / this.scale) - i;
      var dj = Math.floor((prevy-this.y1) / this.scale) - j;
      while (di != 0 || dj != 0) {
        if (-1 < i + di && i + di < (tilesize-1)
         && -1 < j + dj && j + dj < (tilesize-1)) {
          // Connect non-adjacent consecutive pixels
          this.clear(i + di, j + dj);
        }
        di -= di ? di / Math.abs(di) : 0;
        dj -= dj ? dj / Math.abs(dj) : 0;
      }
    }
  };

  tile.prototype.inBounds = function(x, y, prevx, prevy) {
    return this.x1 < x && x < this.x1 + tilesize * this.scale
        && this.y1 < y && y < this.y1 + tilesize * this.scale;
  };

  tile.prototype.cursor = function(ctx, x, y, c, dirty) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    if (!dirty && this.cursi == i && this.cursj == j) {
      return;
    }
    this.cursi = i;
    this.cursj = j;
    this.drawPixel(ctx, x, y, c);
    if (!dirty) {
      this.dirty = true;
    }
  };

  tile.prototype.clearCursor = function() {
    this.cursi = -1;
    this.cursj = -1;
    this.dirty = true;
  }

  tile.prototype.drawPixel = function(ctx, x, y, c) {
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

  tile.prototype.stroke = function(ctx) {
    ctx.lineWidth = 1;
    if (this.active) {
      ctx.strokeStyle = "rgba(255,0,0,1)";
    } else {
      ctx.strokeStyle = "rgba(0,0,0,1)";
    }
    ctx.strokeRect(this.x1, this.y1, this.scale * this.px.length, this.scale * this.px.length);
  }

  return tile

})(Game);

