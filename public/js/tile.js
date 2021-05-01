Game.Tile = (function(g){
  "use strict";

  var tilesize = 16;
  var editLimit = 64;

  var tile = function(imgData, palette){
    this.x1 = 0;
    this.y1 = 0;
    this.scale = 1;
    this.px = [];
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
    if (dirty || this.dirty) {
      this.x1 = x1;
      this.y1 = y1;
      this.scale = scale;
      this.dirty = false;
      this.cursorX = 0;
      this.cursorY = 0;
      var prev = "";
      for (var i = 0; i < this.w; i++) {
        for (var j = 0; j < this.h; j++) {
          if (this.buffer[i][j] != "") {
            ctx.fillStyle = this.buffer[i][j];
            ctx.fillRect(x1 + i * scale, y1 + j * scale, scale, scale);
            continue
          }
          if (this.px[i][j] != prev) {
            ctx.fillStyle = this.px[i][j];
          }
          ctx.fillRect(x1 + i * scale, y1 + j * scale, scale, scale);
        }
      }
    }
  };

  tile.prototype.lock = function() {
    var self = this;
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        self.active = true;
        self.dirty = true;
        resolve();
      }, 100);
    });
  };

  tile.prototype.rollback = function() {
    var self = this;
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
      }, 100);
    });
  };

  tile.prototype.commit = function() {
    var self = this;
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const mask = new BitSet();
        const colors = [];
        var n = 0;
        for (var i in self.buffer) {
          for (var j in self.buffer[i]) {
            n = i*self.w + parseInt(j);
            if (self.buffer[i][j] != "") {
              mask.set(n, 1);
              colors.push(this.palette.getIdx(self.buffer[i][j]));
              self.px[i][j] = self.buffer[i][j];
              self.buffer[i][j] = "";
            }
          }
        }
        self.active = false;
        self.dirty = true;
        self.bufferCount = 0;
        if (colors.length > 0) {
          console.log(hex2b64(mask.slice(0,255).toString(16)), mask.slice(0,255), colors);
        }
        resolve();
      }, 100);
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

  tile.prototype.setXY = function(x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    this.set(i, j, c);
  };

  tile.prototype.clear = function(i, j) {
    if (this.buffer[i][j] == "") {
      return;
    }
    this.buffer[i][j] = "";
    this.bufferCount--;
    this.dirty = true;
  };

  tile.prototype.clearXY = function(x, y) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    this.clear(i, j);
  };

  tile.prototype.inBounds = function(x, y) {
    return this.x1 < x && x < this.x1 + tilesize * this.scale &&
      this.y1 < y && y < this.y1 + tilesize * this.scale
  };

  tile.prototype.cursor = function(ctx, x, y, c) {
    this.drawPixel(ctx, x, y, c);
    this.dirty = true;
  };

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

