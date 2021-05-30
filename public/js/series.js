Game.Series = (function(g){
  "use strict";

  var list = [];
  var stores = {};

  var boardStore = function(boardId) {
    var storeName = "board-"+boardId.toString(16).padStart(4, 0);
    if (!(storeName in stores)) {
      stores[storeName] = localforage.createInstance({name: "Game", storeName: storeName});
    }
    return stores[storeName];
  }

  var series = function(data){
    g.object.extend(this, data);
  };

  series.init = function(allData){
    list = [];
    for (let s of allData) {
      list.push(new series(s));
    }
    return list;
  };

  series.list = function() {
    return list;
  };

  series.findActiveBoard = function(boardId) {
    return new Promise((res, rej) => {
      for (let s of list) {
        for (let b of s.boards) {
          if ((boardId == 0 && b.act) || b.id == boardId) {
            const palette = new Game.Palette(s.palette);
            const board = new Game.Board(Game, boardStore(b.id), b, palette, function() {
              res(board);
            });
            return;
          }
        }
      }
      rej();
    });
  };

  return series

})(Game);
