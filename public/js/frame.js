Game.Frame = (function(g){
  "use strict";

  var frame = function(palette, buffer){
    this.mask = new BitSet();
    this.colors = [];
    this.colorsUniq = {};
    this.colorCount = 0;
    var colorNum = 0;
    var n = 0;
    var w = buffer.length;
    for (var i in buffer) {
      for (var j in buffer[i]) {
        if (buffer[i][j] != "") {
          n = i*w + parseInt(j);
          this.mask.set(n, 1);
          colorNum = palette.getIdx(buffer[i][j])
          this.colors.push(colorNum);
          if (!this.colorsUniq[colorNum]) {
            this.colorCount++;
          }
          this.colorsUniq[colorNum] = true;
        }
      }
    }
    if (this.colors.length > 0) {
      var m = this.mask.slice(0,255);
      console.log(typeof m, typeof m.data, m.toArray(), m.slice(0,255), this.colorCount, this.colors);
    }
  };

  return frame;

})(Game);
