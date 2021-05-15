Game.Frame = (function(g){
  "use strict";

  const headerflag_useMask = 28;
  const headerflag_deleted = 29;
  const headerflag_runLengthEncodedMask = 30;
  const headerflag_runLengthEncodedColorTable = 31;

  var frame = function(tile){
    this.mask = new BitSet();
    this.prev = [];
    this.colors = [];
    this.colorsUniq = {};
    this.colorCount = 0;
    this.userid = 0;
    this.deleted = false;
    this.data = null;
    this.hash = null;
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

  // Touchy
  frame.prototype.toBytes = function() {
    if (this.data) {
      return this.data
    }
    var o = 0;
    var b = new BitSet();
    var bs = n => b.set(o++, parseInt(n));
    var append = (bits, a) => [...a.toString(2).padStart(bits, 0)].forEach(bs);
    append(16, this.userid);
    append(8, this.ti*16 + this.tj);
    append(4, 0 + (this.colorCount-1));
    append(1, 0 + (this.colors.length >= 32)); // headerflag_useMask
    append(1, 0 + (this.deleted)); // headerflag_deleted
    append(1, 0); // headerflag_runLengthEncodedMask
    append(1, 0); // headerflag_runLengthEncodedColorTable

    var bits = Math.ceil(Math.log2(this.colorCount));

    // Run length encoding may produce a smaller color table than simple enumeration
    var n = 0;
    var run = 0;
    for (i in this.colors) {
      if (i == 0 || this.colors[i] != this.colors[i-1] || run == 16) {
        n++;
        run = 0;
      }
      run++;
    }
    if (bits > 0 && n * (4 + bits) < this.colors.length * bits) {
      b.set(headerflag_runLengthEncodedColorTable, 1);
    }

    // Run length encoding may produce a smaller mask than simple enumeration
    var uniq = 0;
    var initial = "1".repeat(16);
    var prev = initial;
    var quads = [];
    for (i = 0; i < 16; i++) {
      quads[i] = "";
      for (var j = 0; j < 16; j++) {
        quads[i] += this.mask.get((i%4)*4 + Math.floor(i/4)*64 + Math.floor(j/4)*16 + j%4);
      }
      if (quads[i] != prev) {
        uniq++;
        prev = quads[i];
      }
    }
    if (16 * uniq + 16 < 256 && 16 * uniq + 16 < this.colors.length*8) {
      b.set(headerflag_runLengthEncodedMask, 1);
    }

    var i = 0;
    if (b.get(headerflag_runLengthEncodedMask)) {
      // Dump entire 256 bit mask to bitset
      prev = initial;
      for (i in quads) {
        if (quads[i] == prev) {
          append(1, 0);
        } else {
          append(1, 1);
          append(16, quads[i]);
        }
        prev = quads[i];
      }
    } else if (b.get(headerflag_useMask)) {
      // Dump entire 256 bit mask to bitset
      for (i = 0; i < 256; i++) {
        bs(this.mask.get(i));
      }
    } else {
      append(8, this.colors.length);
      // Dump 8 bit pixel positions to bitset
      this.mask.toArray().forEach(a => append(8, a));
    }

    var runLengthEncode = function(colors, cm) {
      n = 0;
      for (i in colors) {
        if (n == 15 || i == colors.length - 1 || colors[i] != colors[parseInt(i)+1]) {
          append(4, n);
          append(bits, cm ? cm[colors[i]] : colors[i]);
          n = 0;
        } else {
          n++;
        }
      }
    };

    if (bits == 0) {
      append(4, this.colors[0]);
    } else if (bits < 4) {
      var c = [];
      var cm = {};
      for (i in this.colors) {
        if (!(this.colors[i] in cm)) {
          cm[this.colors[i]] = c.length;
          c.push(this.colors[i]);
        }
      }
      // Color index
      for (i in c) {
        append(4, c[i]);
      }
      if (b.get(headerflag_runLengthEncodedColorTable)) {
        // Run length encoded pixel color indices
        runLengthEncode(this.colors, cm);
      } else {
        // Enumerated pixel color indices
        for (i in this.colors) {
          append(bits, cm[this.colors[i]]);
        }
      }
    } else {
      if (b.get(headerflag_runLengthEncodedColorTable)) {
        // Run length encoded pixel colors
        runLengthEncode(this.colors);
      } else {
        // Enumerated pixel colors
        for (i in this.colors) {
          append(bits, this.colors[i]);
        }
      }
    }
    b = b.slice(0, o);
    this.data = (new Int32Array(b.data)).buffer.slice(0, Math.ceil(o/8));
    return this.data;
  };

  frame.prototype.getHash = function() {
    if (this.hash) {
      return Promise.resolve(this.hash);
    }
    var self = this;
    return crypto.subtle.digest('SHA-256', this.toBytes().slice(2)).then(h => {
      self.hash = hex2b64((new Uint8Array(h)).reduce((a, c) => a += c.toString(16).padStart(2, '0'), ''));
      return self.hash;
    });
  };

  frame.fromBytes = function(bytes) {
    var b = new BitSet(new Uint8Array(bytes));
    var i = 0;
    var h = b.slice(0, 31);
    var f = new Game.Frame();
    f.data = bytes;
    f.userid = intAt(b, 16, 0);
    var tileID = intAt(b, 8, 16);
    f.ti = Math.floor(tileID/16);
    f.tj = tileID % 16;
    f.colorCount = intAt(b, 4, 24)+1;
    f.deleted = !!b.get(headerflag_deleted);
    var useMask = b.get(headerflag_useMask);
    var o = 32;
    var bits = Math.ceil(Math.log2(f.colorCount));
    var numpx = 0;
    if (b.get(headerflag_runLengthEncodedMask)) {
      var initial = "1".repeat(16);
      var quad = initial;
      for (i = 0; i < 16; i++) {
        if (intAt(b, 1, o)) {
          quad = intAt(b, 16, o+1).toString(2).padStart(16, 0);
          o += 16;
        }
        o++;
        for (var j = 0; j < 16; j++) {
          f.mask.set((i%4)*4 + Math.floor(i/4)*64 + Math.floor(j/4)*16 + j%4, parseInt(quad[j]));
          numpx += parseInt(quad[j]);
        }
      }
    } else if (b.get(headerflag_useMask)) {
      f.mask = b.slice(o, o + 255);
      numpx = f.mask.cardinality();
      o += 256;
    } else {
      numpx = intAt(b, 8, o);
      o += 8;
      for (i = 0; i < numpx; i++) {
        f.mask.set(intAt(b, 8, o), 1);
        o += 8;
      }
    }

    // Decode color index (if exists)
    var cm = {};
    if (bits < 4) {
      for (i = 0; i < f.colorCount; i++) {
        cm[i] = intAt(b, 4, o);
        o += 4;
      }
    }
    // Decode color table
    var c;
    if (b.get(headerflag_runLengthEncodedColorTable)) {
      var n;
      for (i = 0; i < numpx; i++) {
        n = intAt(b, 4, o);
        c = intAt(b, bits, o + 4);
        f.colors.push(...Array(n+1).fill(bits < 4 ? cm[c] : c));
        o += bits + 4;
        i += n;
      }
    } else {
      for (i = 0; i < numpx; i++) {
        c = bits > 0 ? intAt(b, bits, o) : 0;
        f.colors.push(bits < 4 ? cm[c] : c);
        o += bits;
      }
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
