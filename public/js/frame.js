Game.Frame = (function(g){
  "use strict";

  const headerflag_useMask = 29;
  const headerflag_deleted = 30;
  const headerflag_runlengthencoding = 31;
  const headerflag_reserved = 32;

  var frame = function(tile){
    this.mask = new BitSet();
    this.prev = [];
    this.colors = [];
    this.colorsUniq = {};
    this.colorCount = 0;
    this.userid = 0;
    if (!tile) {
      return
    }
    this.ti = tile.ti;
    this.tj = tile.tj;
    var colorNum = 0;
    var n = 0;
    var w = tile.buffer.length;
    for (var i in tile.buffer) {
      for (var j in tile.buffer[i]) {
        if (tile.buffer[i][j] != "") {
          this.mask.set(i*w + parseInt(j), 1);
          colorNum = tile.palette.getIdx(tile.buffer[i][j]);
          this.colors.push(colorNum);
          this.prev.push(tile.palette.getIdx(tile.px[i][j]));
          if (!this.colorsUniq[colorNum]) {
            this.colorCount++;
          }
          this.colorsUniq[colorNum] = true;
        }
      }
    }
    this.mask = this.mask.slice(0,255);
  };

  // Look. This is complicated. ONLY LOOK! NO TOUCH!
  frame.prototype.toBytes = function() {
    var o = 0;
    var b = new BitSet();

    [...this.userid.toString(2).padStart(16, 0)].forEach(n => b.set(o++, parseInt(n)));
    [...this.ti.toString(2).padStart(4, 0)].forEach(n => b.set(o++, parseInt(n)));
    [...this.tj.toString(2).padStart(4, 0)].forEach(n => b.set(o++, parseInt(n)));
    [...(this.colorCount-1).toString(2).padStart(4, 0)].forEach(n => b.set(o++, parseInt(n)));
    b.set(headerflag_useMask, this.colors.length >= 32);
    b.set(headerflag_deleted, this.deleted);
    b.set(headerflag_runlengthencoding, 0);
    b.set(headerflag_reserved, 0);
    o += 4;

    var i = 0;
    if (b.get(headerflag_useMask)) {
      // Dump entire 256 bit mask to bitset
      for (i = 0; i < 256; i++) {
        b.set(o++, this.mask.get(i));
      }
    } else {
      // Dump 8 bit pixel positions to bitset
      var a = this.mask.toArray();
      for (i in a) {
        [...a[i].toString(2).padStart(8, 0)].forEach(n => b.set(o++, parseInt(n)));
      }
    }

    // We don't need 4 bits per color if we only have 2 colors.
    var bits = Math.ceil(Math.log2(this.colorCount));
    if (bits < 4 && bits > 0) {
      var c = [];
      var cm = {};
      for (i in this.colors) {
        if (!cm[this.colors[i]]) {
          cm[this.colors[i]] = c.length;
          c.push(this.colors[i]);
        }
      }
      // Color index
      for (i in c) {
        [...c[i].toString(2).padStart(4, 0)].forEach(n => b.set(o++, parseInt(n)));
      }
      // Pixel color indices
      for (i in this.colors) {
        [...cm[this.colors[i]].toString(2).padStart(bits, 0)].forEach(n => b.set(o++, parseInt(n)));
      }
    } else if (bits == 0) {
        [...this.colors[0].toString(2).padStart(4, 0)].forEach(n => b.set(o++, parseInt(n)));
    } else {
      for (i in this.colors) {
        [...this.colors[i].toString(2).padStart(4, 0)].forEach(n => b.set(o++, parseInt(n)));
      }
    }
    return b.slice(0, o);
  };

  frame.fromBytes = function(b) {
    var i = 0;
    var h = b.slice(0, 31);
    var f = new Game.Frame();
    f.userid = intAt(b, 16, 0);
    f.ti = intAt(b, 4, 16);
    f.tj = intAt(b, 4, 20);
    f.colorCount = intAt(b, 4, 24)+1;
    f.deleted = !!b.get(headerflag_deleted);
    var useMask = b.get(headerflag_useMask);
    var o = 32;
    var hex = b.slice(o).toString(16);
    var bits = Math.ceil(Math.log2(f.colorCount));
    if (useMask) {
      f.mask = b.slice(o, o + 255);
      numpx = f.mask.cardinality();
      o += 256;
    } else {
      var numpx = 0;
      if (f.colorCount == 1) {
        numpx = Math.ceil((hex.length - 1) / 2);
      } else if (bits == 4) {
        numpx = Math.floor((hex.length) / 2 / (1 + bits/8));
      } else {
        numpx = Math.floor((hex.length - f.colorCount) / 2 / (1 + bits/8));
      }
      for (i = 0; i < numpx; i++) {
        f.mask.set(intAt(b, 8, o), 1);
        o += 8;
      }
    }
    var cm = {};
    if (bits < 4) {
      for (i = 0; i < f.colorCount; i++) {
        cm[i] = intAt(b, 4, o);
        o += 4;
      }
    }
    var c;
    for (i = 0; i < numpx; i++) {
      c = bits > 0 ? intAt(b, bits, o) : 0;
      f.colors.push(bits < 4 ? cm[c] : c);
      o += bits;
    }
    f.mask = f.mask.slice(0,255);
    return f;
  };

  var j;
  var intAt = function(b, bits, pos) {
    var n = 0;
    for (j = 0; j < bits; j++) {
      n = n << 1;
      n += b.get(pos + j);
    }
    return n;
  }

  return frame;

})(Game);
