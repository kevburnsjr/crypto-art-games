Game.Test = (function(g){
  "use strict";

  var test = function(){};

  test.random = async function(n) {
    const t = new Game.Tile(null, Game.board().palette, 0, 0);
    var data = [];
    var f;
    const s = Game.getSocket();
    for (var i = 0; i < n; i++) {
      t.ti = Math.floor(i/16)%16;
      t.tj = i%16
      for (var j = 0; j < 256; j++) {
        if (Math.floor(Math.random() * 16) > 0) {
          t.buffer[Math.floor(j/16)][j%16] = Math.floor(Math.random() * 16);
        } else {
          t.buffer[Math.floor(j/16)][j%16] = null;
        }
      }
      f = new Game.Frame(t);
      await s.sendFrame(f);
    }
  };

  return test;

})(Game);
