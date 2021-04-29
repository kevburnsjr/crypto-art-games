Game.Tile = (function(g){
  "use strict";

  var tilesize = 16;

  var tile = function(imgData){
    this.x1 = 0;
    this.y1 = 0;
    this.scale = 1;
    this.px = [];
    this.active = false;
    for (var i = 0; i < imgData.width; i++) {
      this.px[i] = [];
      for (var j = 0; j < imgData.height; j++) {
        var n = j*imgData.width*4 + i*4;
        this.px[i][j] = "rgb("+imgData.data[n]+","+imgData.data[n+1]+","+imgData.data[n+2]+")";
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

  tile.prototype.deactivate = function() {
    this.active = false;
    this.dirty = true;
  };

  tile.prototype.toggleActive = function() {
    this.active = !this.active;
    this.dirty = true;
  };

  tile.prototype.get = function(i, j) {
    return this.px[i][j]
  };

  tile.prototype.getXY = function(x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    return this.get(i, j);
  };

  tile.prototype.set = function(i, j, c) {
    this.px[i][j] = c;
    this.dirty = true;
  };

  tile.prototype.setXY = function(x, y, c) {
    var i = Math.floor((x-this.x1) / this.scale);
    var j = Math.floor((y-this.y1) / this.scale);
    this.set(i, j, c);
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

