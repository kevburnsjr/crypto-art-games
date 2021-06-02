var Game = (function(g){
  "use strict";

  var animationFrame;
  var bgcolor = "#666";
  var defaultZoom = 3;
  var zoom = defaultZoom;
  var prevZoom = zoom;
  var tile = 0;
  var focused = false;
  var color = Math.floor(Math.random() * 16);
  var bgCtx, bgElem;
  var uiCtx, uiElem;
  var dirty = false;
  var uiDirty = false;
  var w;
  var h;
  var allSeries = [];
  var board;
  var nav;
  var hoverX;
  var hoverY;
  var bgtimeout = null;
  var socket;
  var policy;
  var boardId = 0;
  var store = {};
  const stores = ["global", "user", "ui", "reports", "bans"];

  var createStores = async function() {
    stores.map((name) => store[name] = localforage.createInstance({name: "Game", storeName: name}));
    return Promise.resolve()
  }

  var start = async function(bgCanvasElem, uiCanvasElem, paletteElem, leftNavElem, rightNavElem, botNavElem, scrubberElem, modalElem) {
    await createStores();
    bgElem = bgCanvasElem;
    bgCtx = bgElem.getContext('2d', { alpha: true });
    uiElem = uiCanvasElem;
    uiCtx = uiElem.getContext('2d', { alpha: true });
    nav = new Game.Nav(Game, store.ui, leftNavElem, rightNavElem, botNavElem, scrubberElem, modalElem);
    document.addEventListener('mousemove', mousemove);
    document.addEventListener('mousedown', mousedown);
    document.addEventListener('mouseup', mouseup);
    document.addEventListener('click', click);
    document.addEventListener('keydown', keydown);
    document.addEventListener('keyup', keyup);
    document.addEventListener('keypress', keypress);
    document.addEventListener('wheel', wheel);
    document.addEventListener('visibilitychange', visibilitychange);
    window.addEventListener('blur', visibilitychange);
    window.addEventListener('resize', resize);
    window.addEventListener('contextmenu', e => e.preventDefault());
    window.addEventListener('paste', paste);
    var userID = null;
    var banIdx    = parseInt(await store.global.getItem("banIdx"), 16) || 0;
    var userIdx   = parseInt(await store.global.getItem("userIdx"), 16) || 0;
    var reportIdx = parseInt(await store.global.getItem("reportIdx"), 16) || 0;
    // initiate websocket
    socket = new Game.socket({
      initializing: false,
      awaiting: null,
      boardChangeCallback: null,
      url: function() {
        return `/socket?banIdx=${banIdx}&userIdx=${userIdx}&reportIdx=${reportIdx}`;
      },
      changeBoard: async function(id, callback) {
        if (socket.initializing || (board && board.id == id)) {
          return Promise.resolve();
        }
        socket.initializing = true;
        socket.boardChangeCallback = callback ? callback : null;
        return Game.Series.findActiveBoard(id).then(async b => {
          board = b;
          boardId = board.id;
          uiDirty = true;
          socket.send(JSON.stringify({
            type:       'board-init',
            boardId:    boardId,
            generation: parseInt(await board.store.getItem("generation"), 16) || 0,
            timecode:   parseInt(await board.store.getItem("timecode"),   16) || 0
          }));
        }).catch((e) => log("Failed to load board", id, e));
      },
      sendFrame: function(f) {
        return new Promise((resolve, reject) => {
          f.getHash().then(function(hash) {
            socket.awaiting = hash;
            socket.on('complete', function(f) {
              f.getHash().then((h) => {
                if (h == hash) {
                  socket.awaiting = null;
                  nav.resetScrubber();
                  resolve(f);
                }
              });
            });
            socket.send(f.data);
          });
        }).finally(() => {
          socket.off('complete');
        });
      },
      undoFrame: function(board, f) {
        var fn;
        return new Promise((resolve, reject) => {
          socket.awaiting = f.timecode;
          fn = function(e) {
            if (e.timecode == socket.awaiting) {
              resolve(e);
            }
          };
          socket.on('frame-undone', fn);
          socket.send(JSON.stringify({type:'frame-undo', boardId: board.id, timecode: f.timecode}));
        }).finally(() => {
          socket.off('frame-undone');
        });
      },
      lockTile: function(t) {
        return new Promise((resolve, reject) => {
          if (userID == null) {
            nav.showLoginModal();
            reject();
            return
          }
          socket.on(['tile-locked', 'err'], function(e) {
            if (e.type == 'err') {
              nav.flash("error", e.msg, 1500);
              reject(e.msg);
            } else if (e.userID == userID){
              nav.showHeart(e.bucket);
              resolve();
            }
          });
          socket.send(JSON.stringify({type:'tile-lock', tileID: t.getID()}));
        }).finally(msg => {
          socket.off('tile-locked');
          socket.off('err');
          return msg;
        });
      },
      unlockTile: function(t) {
        return new Promise((resolve, reject) => {
          socket.on(['tile-lock-released', 'err'], function(e) {
            if (e.type == 'err') {
              nav.flash("error", e.msg, 1500);
              nav.resetScrubber();
              reject(e.msg);
            } else if (e.userID == userID) {
              nav.showHeart(e.bucket);
              nav.resetScrubber();
              resolve();
            }
          });
          socket.send(JSON.stringify({type:'tile-lock-release', tileID: t.getID()}));
        }).finally(msg => {
          socket.off('tile-lock-released');
          socket.off('err');
          return msg;
        });
      },
      report: async function(timecode, reason) {
        var f = board.frames[timecode];
        const user = await Game.User.find(f.userid);
        return new Promise((resolve, reject) => {
          if (userID == null) {
            nav.showLoginModal();
            reject();
            return
          }
          socket.once(['report-success'], function(e) {
            if (e.userID == userID){
              nav.flash("success", `${user.display_name} reported for ${reason}`, 1500);
              resolve();
            }
          });
          socket.send(JSON.stringify({type:'report', boardId: board.id, date: f.date, timecode: parseInt(timecode), reason: reason}));
        });
      },
      clearReports: async function(targetID) {
        return new Promise((resolve, reject) => {
          socket.send(JSON.stringify({type:'report-clear', targetID: parseInt(targetID)}));
        });
      },
      userBan: async function(targetID, date, duration) {
        return new Promise((resolve, reject) => {
          socket.send(JSON.stringify({type:'user-ban', targetID: parseInt(targetID), since: parseInt(date), duration: duration}));
        });
      },
      love: async function(timecode) {
        var f = board.frames[timecode];
        const user = await Game.User.find(f.userid);
        return new Promise((resolve, reject) => {
          if (userID == null) {
            nav.showLoginModal();
            reject();
            return
          }
          socket.once(['love'], function(e) {
            if (e.userID == userID){
              nav.flash("success", `Liked contribution from ${user.display_name}`, 1500);
              resolve();
            }
          });
          socket.send(JSON.stringify({type:'love', boardId: board.id, timecode: parseInt(timecode)}));
        });
      },
      errStorage: async function() {
        return new Promise((resolve, reject) => {
          socket.send(JSON.stringify({type:'err-storage', userID: userID || 0, userAgent: navigator.userAgent}));
        });
      },
    });
    socket.on('message', function(msg) {
      if (msg instanceof ArrayBuffer) {
        const f = Game.Frame.fromBytes(msg);
        return board.saveFrame(f).then(() => {
          if (socket.awaiting) {
            socket.emit('complete', f);
          }
        });
      } else {
        var e = JSON.parse(msg);
        return socket.serial(e.type, e);
      }
    });
    socket.on('board-init-complete', async (e) => {
      if (board != null) {
        nav.showHeart(e.bucket);
        await board.enable(e.timecode);
        socket.initializing = false;
      }
      if (socket.boardChangeCallback != null) {
        socket.boardChangeCallback(board);
        socket.boardChangeCallback = null;
      }
    });
    socket.on('new-user', function(e) {
      const user = new Game.User(e);
      user.save();
    });
    socket.on('logout', function(e) {
      window.location.href = "/logout";
    });
    socket.on('report', function(e) {
      store.reports.setItem([e.targetID, e.boardID, e.timecode, e.userID].join("-"), e);
      nav.showMod();
    });
    socket.on('report-clear', async function(e) {
      var toRemove = [];
      await store.reports.iterate((v, k, i) => {
        if (k.split("-")[0] == e.targetID) {
          toRemove.push(k);
        }
      });
      for (let k of toRemove) {
        await store.reports.removeItem(k);
      }
      nav.showMod();
    });
    socket.on('user-ban', async function(e) {
      const board = Game.board();
      var bans;
      var boardStore;
      for (let s of Game.Series.list()) {
        for (let b of s.boards) {
          if (board && board.id == b.id) {
            await board.applyUserBan(e);
            nav.showRecent(board);
            continue;
          }
          boardStore = Game.Series.boardStore(b.id);
          bans = await boardStore.getItem("_bans");
          if (bans == null) {
            bans = [];
          }
          bans.push(e);
          await boardStore.setItem("_bans", bans);
        }
      }

      nav.showMod();
    });
    socket.on('init', async function(e) {
      return new Promise((resolve, reject) => {
        checkVersion(e.v).catch((d) => {
          if (d == undefined) {
            log("New version ", e.v);
            window.location.reload();
          } else {
            log(d);
          }
        });
        socket.initializing = false;
        if (e.user) {
          userID = e.user.userID;
          policy = e.user.policy;
        }
        // processBans(banIdx, e.banIdx);
        nav.showMod();
        store.global.setItem("banIdx", e.banIdx.toString(16).padStart(4, 0));
        store.global.setItem("userIdx", e.userIdx.toString(16).padStart(4, 0));
        store.global.setItem("reportIdx", e.reportIdx.toString(16).padStart(4, 0));
        banIdx = e.banIdx;
        userIdx = e.userIdx;
        reportIdx = e.reportIdx;
        nav.init(e.user);
        if (e.user && e.user.id != null) {
          if(!policy) {
            nav.showPolicyModal();
            return;
          }
        }
        if (e.user && e.user.mod) {
          document.body.classList.add("mod")
        }
        if (!e.series) {
          log("Series missing from init", e);
          return;
        }
        Game.Series.init(e.series);
        nav.showSeries(Game.Series.list());
        socket.changeBoard(boardId, board => {
          if (focused) {
            board.setFocus(Math.floor(tile/16), tile%16);
          }
        });
        resolve();
      });
    });
    reset();
    await socket.start();
  };

  var checkVersion = function (v) {
    return new Promise((res, rej) => {
      var done = false;
      store.global.getItem("_v").then((_v) => {
        if (v === undefined || v === null) {
          done = true;
          res();
          return;
        }
        if (_v === null) {
          store.global.setItem("_v", v).then(res);
        } else if (_v !== v) {
          localforage.dropInstance({name: "Game"})
            .then(createStores)
            .then(() => store.global.setItem("_v", v))
            .then(rej);
        } else {
          res();
        }
        done = true;
      }).catch(log);
      setTimeout(() => {
        if (!done) {
          socket.errStorage();
        }
      }, 1000);
      // IndexedDB frequently becomes corrupt.
      // Reads hang forever so detecting unrespon
    });
  }

  var reset = function() {
    setZoom();
    w = window.innerWidth;
    h = window.innerHeight;
    if(window.location.hash) {
      const parts = window.location.hash.substr(1).split(':');
      if (parts.length > 0) boardId = parseInt(parts[0]);
      if (parts.length > 1) tile    = parseInt(parts[1]);
      if (parts.length > 2) color   = parseInt(parts[2]);
      if (parts.length > 3) zoom    = Math.max(parseInt(parts[3] != undefined ? parts[3] : 0), 1);
      if (parts.length > 4) focused = parts[4] == "1";
      if (board && board.id != boardId) {
        socket.changeBoard(boardId, (board) => {
          // console.log(Math.floor(tile/16), tile%16);
          board.setFocus(Math.floor(tile/16), tile%16);
        });
      } else if (board) {
        if (focused) {
          board.setFocus(Math.floor(tile/16), tile%16);
        } else {
          board.cancelFocus();
        }
      }
    }
    window.cancelAnimationFrame(animationFrame);
    draw();
    bgElem.style.display = "block";
  };

  var prevBoardId;

  var draw = function() {
    if (board == null) {
      animationFrame = window.requestAnimationFrame(draw);
      return;
    }
    if (bgElem.width != w || bgElem.height != h || zoom != prevZoom) {
      bgElem.width = w;
      bgElem.height = h;
      uiElem.width = w;
      uiElem.height = h;
      prevZoom = zoom;
      dirty = true;
      uiDirty = true;
    }
    if (dirty || board.dirty) {
      bgCtx.clearRect(0, 0, w, h);
    }
    if (uiDirty || board.uiDirty) {
      uiCtx.clearRect(0, 0, w, h);
    }
    try {
      board.render(bgCtx, uiCtx, w/2, h/2, hoverX, hoverY, zoom, dirty, uiDirty, mousedown);
    } catch(e) {
      bgtimeout = setTimeout(function(){
        draw();
      }, 1000);
      return;
    }
    dirty = false;
    uiDirty = false;
    animationFrame = window.requestAnimationFrame(draw);
  };

  var setTimecode = function(tc) {
    board.timecode = Math.max(tc, 0);
  };

  // ----------------- Input Functions -------------------

  var clickpoint = [];
  var isMousedown = false;
  var worldnav = false;
  var brushState = false;
  var keyDownMap = {};
  var isKeyDown = function() {
    for (let k of arguments) {
      if (keyDownMap[k]) {
        return true;
      }
    }
    return false;
  }

  // click
  var click = function(e){
    if (e.target.id == "uicanvas") {
      board.handleClick(e, w/2, h/2, hoverX, hoverY, zoom);
      setHash();
    }
  };

  // mousedown
  var mousedown = function(e){
    var t = e.target;
    if (t.parentElement.id == "logo") {
      e.preventDefault();
      document.getElementById("world-nav").classList.add("open");
      worldnav = true;
      return
    }
    if (t.id == "brush-state") {
      e.preventDefault();
      if (!board.palette.active) {
        board.palette.showBottom();
        brushState = true;
      }
      return
    }
    isMousedown = true;
    clickpoint = [e.offsetX, e.offsetY];
    if (e.target.nodeName == "CANVAS" && t.parentNode.id == "palette") {
      e.preventDefault();
      e.stopPropagation();
      board.palette.setXY(e.pageX, e.pageY);
      setColor();
      if (brushState) {
        board.palette.hide();
        brushState = false;
      }
    } else if (board.palette.active) {
      e.preventDefault();
      if (t.parentNode.parentNode.id != "palette") {
        board.palette.hide();
        brushState = false;
      }
      return;
    }
    board.handleMouseDown(w/2, h/2, hoverX, hoverY);
  };

  // mousemove
  var mousemove = function(e){
    if (board == null) {
      return;
    }
    hoverX = Math.round(e.pageX);
    hoverY = Math.round(e.pageY);
    board.handleMouseMove(hoverX, hoverY, isMousedown, board.palette.getColor());
  };

  // mouseup
  var mouseup = function(e){
    isMousedown = false;
    document.body.classList.remove("reporting");
    if (worldnav) {
      document.getElementById("world-nav").classList.remove("open");
      worldnav = false;
      return
    }
    if (brushState && e.target.nodeName == "CANVAS" && e.target.parentNode.id == "palette") {
      e.preventDefault();
      e.stopPropagation();
      board.palette.setXY(e.pageX, e.pageY);
      setColor();
      board.palette.hide();
      brushState = false;
    }
    board.clearPath();
    clickpoint = [];
  };

  // keydown
  var keydown = function(e){
    var k = e.key.toLowerCase();
    keyDownMap[k] = true;
    if (k == "alt") {
      e.preventDefault();
      if (!document.body.classList.contains("color-picking")) {
        document.body.classList.add("color-picking");
        board.toggleDropper();
      }
    }
    if (k == "e") {
      e.preventDefault();
      if (!document.body.classList.contains("erasing")) {
        document.body.classList.add("erasing");
        board.toggleEraser();
      }
    }
    if (k == "tab") {
      e.preventDefault();
      e.stopPropagation();
      if (!board.palette.active) {
        board.palette.show(hoverX, hoverY);
        board.togglePalette();
      }
    }
    if (k == "w" || k == "arrowup") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(0, -1).then(setHash);
      document.body.classList.remove("editing");
    }
    if (k == "a" || k == "arrowleft") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(-1, 0).then(setHash);
      document.body.classList.remove("editing");
    }
    if (k == "s" || k == "arrowdown") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(0, 1).then(setHash);
      document.body.classList.remove("editing");
    }
    if (k == "d" || k == "arrowright") {
      e.preventDefault();
      // if ctrl move boards else move tiles
      board.moveTile(1, 0).then(setHash);
      document.body.classList.remove("editing");
    }
    if (k == "0" || k == "numpad0") {
      zoom = 3;
      board.cancelFocus();
      setZoom();
      setHash();
    }
    if (k == "[" || k == "-") {
      board.setBrushSize(0);
    }
    if (k == "]" || k == "=") {
      board.setBrushSize(1);
    }
    if (k == "pageup") {
      e.preventDefault();
      // Navigate to next board
    }
    if (k == "pagedown") {
      e.preventDefault();
      // Navigate to previous board
    }
    if (k == "escape") {
      e.preventDefault();
      if (nav.handleEscape()) {
        return
      }
      if (brushState) {
        board.palette.hide();
        brushState = false;
        return
      }
      if (board.tile && board.tile.active) {
        board.cancelActive().then(function() {
          document.body.classList.remove("editing");
        });
        return
      } else if (board.focused) {
        board.cancelFocus();
      }
    }
  };

  // keyup
  var keyup = function(e){
    var k = e.key.toLowerCase();
    keyDownMap[k] = false;
    if (k == "alt") {
      document.body.classList.remove("color-picking");
      board.toggleDropper();
    }
    if (k == "e") {
      document.body.classList.remove("erasing");
      board.toggleEraser();
    }
    if (k == "tab") {
      e.preventDefault();
      e.stopPropagation();
      board.palette.hide();
      board.togglePalette();
    }
  };

  // keypress
  var keypress = function(e){
    var k = e.key.toLowerCase();
    if (e.key == " ") {
      e.preventDefault();
      board.toggleActive().then(function(active){
        if (active) {
          document.body.classList.add("editing");
        } else {
          document.body.classList.remove("editing");
        }
      });
    }
    if (k == "h") {
      e.preventDefault();
      nav.toggleHelp();
    }
    if (k == "b") {
      e.preventDefault();
      nav.toggleSeries();
    }
    if (k == "r") {
      e.preventDefault();
      nav.toggleRecentFrames();
    }
    if (k == "c") {
      e.preventDefault();
      nav.toggleChat();
    }
  };

  // wheel
  var wheel = function(e) {
    if (e.shiftKey) {
      Game.nav().handleWheel(e);
      return;
    }
    if (e.deltaY < 0) {
      if (zoom < 6) {
        zoom += 1;
      } else if (zoom < 12) {
        zoom += 2;
      } else {
        zoom += 4;
      }
    }
    if (e.deltaY > 0) {
      if (zoom <= 6) {
        zoom -= 1;
      } else if (zoom <= 12) {
        zoom -= 2;
      } else {
        zoom -= 4;
      }
    }
    setZoom();
    setHash();
  };

  // ----------------- View Functions -------------------

  var resize = function(e){
    clearTimeout(bgtimeout);
    bgtimeout = setTimeout(function(){
      w = window.innerWidth;
      h = window.innerHeight;
    }, 100);
  };

  var visibilitychange = function(e){
    document.getElementById("world-nav").classList.remove("open");
    document.body.classList.remove("color-picking", "erasing", "editing", "reporting");
    keyDownMap = {};
    if (board != null) {
      board.cancelActive();
    }
  };

  // ----------------- Clipboard  Functions -------------------

  var paste = function(e) {
    if (!board.tile || !board.tile.active) {
      return
    }
    var items = (e.clipboardData  || e.originalEvent.clipboardData).items;
    var blob = null;
    for (var i = 0; i < items.length; i++) {
      if (items[i].type.indexOf("image") === 0) {
        blob = items[i].getAsFile();
      }
    }
    if (blob !== null) {
      var reader = new FileReader();
      reader.onload = function(e) {
        var icanvas = document.createElement('canvas');
        var ictx = icanvas.getContext("2d");
        var img = new Image();
        img.onload = function() {
          icanvas.width = board.tileSize;
          icanvas.height = board.tileSize;
          ictx.drawImage(img, 0, 0, board.tileSize, board.tileSize);
          board.tile.setBufferData(ictx.getImageData(0,0,board.tileSize,board.tileSize));
        };
        img.src = e.target.result;
      };
      reader.readAsDataURL(blob);
    }
  }

  // ----------------- State Functions -------------------

  var setHash = function() {
    window.location.assign(window.location.href.split("#")[0] + "#" + [
      board ? board.id : boardId,
      board ? board.getTileID() : tile,
      board ? board.palette.color : color,
      zoom,
      (board ? board.focused : focused)?1:0,
    ].join(':'));
  };

  window.onhashchange = function() {
    reset();
  };

  var setZoom = function() {
    zoom = Math.max(1, Math.min(32, zoom));
  };

  var setColor = function() {
    document.getElementById("brush-state").style.display = "block";
    document.getElementById("brush-state").style.backgroundColor = board.palette.getColor();
  };

  var log = function() {
    if ("console" in window) {
      console.trace(...arguments);
    }
  };

  var getSocket = function() {
    return socket;
  };

  return {
    start: start,
    mousedown: function(){
      return mousedown;
    },
    setColor: setColor,
    isKeyDown: isKeyDown,
    online: false,
    log: log,
    nav: () => nav,
    setTimecode: setTimecode,
    getSocket: getSocket,
    board: () => board,
    store: () => store
  };

})({});
