Game.Nav = (function(g){
  "use strict";

  var nav = function(game, uiStore, left, right, bot, scrubber, modal){
    this.initialized = false;
    this.game = game;
    this.timeagoInterval = null;
    this.toggles = {};
    this.scrubber = scrubber;
    this.modal = modal;
    this.showRecentAnimationFrame = null;
    this.recentFrames = document.getElementById("recent-frames");
    this.reportEl = document.getElementById("report");
    this.modEl = document.getElementById("mod");
    this.flashEl = document.getElementById("flash");
    this.recentTileTpl = document.getElementById("recent-frames-li").innerHTML;
    this.flashTimeout = null;
    this.heartTimeout = null;
    this.recentTiles = [];
    this.demoMode = localStorage.getItem('demo-mode') == "true";
    this.demoInterval = null;
    this.demoState = {};

    const toggleGroup = async el => {
      const toggles = el.querySelectorAll("nav a");
      const panels = el.querySelectorAll("div");
      el.querySelector("nav").addEventListener('click', async e => {
        e.preventDefault();
        var $a = g.dom.matchParent(e.target, "a");
        if (!$a) {
          return;
        }
        var active = $a.classList.contains('active');
        toggles.forEach(async t => {
          var id = t.dataset.toggle;
          await uiStore.setItem("ui-"+id, false);
          t.classList.remove('active');
          document.getElementById(id).classList.remove('active');
        })
        if (!active) {
          var id = $a.dataset.toggle;
          await uiStore.setItem("ui-"+id, true);
          $a.classList.add('active');
          document.getElementById(id).classList.add('active');
        }
      });
      toggles.forEach((t, i) => {
        var id = t.dataset.toggle;
        this.toggles[id] = t
        uiStore.getItem("ui-"+id).then(active => {
          if (i == 0 && active == null) {
            uiStore.setItem("ui-"+id, true);
            active = true
          }
          if (active) {
            t.click();
          }
        })
      });
    };
    toggleGroup(left);
    toggleGroup(right);
    toggleGroup(bot);

    this.handleWheel = e => {
      e.stopPropagation();
      if (game.board().tile.active) {
        return;
      }
      scrubber.scrollLeft += e.deltaY/Math.abs(e.deltaY);
    };
    scrubber.addEventListener('scroll', e => {
      e.stopPropagation();
      if (game.board().tile.active) {
        e.preventDefault();
        return;
      }
      game.board().setOffset(scrubber.scrollWidth - scrubber.offsetWidth - Math.round(scrubber.scrollLeft*window.devicePixelRatio));
    });
    this.recentFrames.addEventListener('wheel', this.handleWheel, { passive: true });
    this.recentFrames.addEventListener('click', e => {
      e.preventDefault();
      if (e.target.nodeName == "CANVAS") {
        game.board().setFocus(parseInt(e.target.dataset.i), parseInt(e.target.dataset.j));
        return;
      }
      var t = e.target;
      if (t.classList.contains("heart")) {
        t = t.parentNode;
      }
      if (t.nodeName == "A" && t.classList.contains("love")) {
        game.getSocket().love(t.parentNode.parentNode.dataset.timecode);
      }
    });
    var self = this;
    this.recentFrames.addEventListener('mousedown', e => {
      e.preventDefault();
      if (e.target.nodeName == "A" && e.target.classList.contains("report")) {
        self.reportEl.style.right = 0;
        self.reportEl.style.top = e.target.getBoundingClientRect().bottom;
        self.reportEl.dataset.timecode = e.target.parentNode.parentNode.dataset.timecode;
        self.reportEl.dataset.userid = e.target.parentNode.parentNode.dataset.userid;
        self.reportEl.dataset.date = e.target.parentNode.parentNode.dataset.date;
        document.querySelectorAll('.reporting').forEach((el) => el.classList.remove('reporting'));
        document.body.classList.add('reporting');
        e.target.parentNode.parentNode.classList.add('reporting');
      }
    });
    this.recentFramesTimeago = function() {
      self.recentFrames.querySelectorAll('.timeago').forEach((el) => {
        el.textContent = timeago.format(el.getAttribute("datetime"));
      });
    };
    setInterval(this.recentFramesTimeago, 5000);
    this.reportEl.addEventListener('mouseup', e => {
      if (document.body.classList.contains('reporting')) {
        const reason = e.target.dataset.reason;
        if (reason == "timeout") {
          self.showTimeoutModal(e.target.parentNode.dataset.userid, e.target.parentNode.dataset.date);
        } else if (reason == "clear") {
          self.showReportClearModal(e.target.parentNode.dataset.userid);
        } else {
          game.getSocket().report(
            e.target.parentNode.dataset.timecode,
            e.target.dataset.reason
          );
        }
      }
    });
    this.reportEl.addEventListener('mouseover', e => {
      if (e.target.dataset.reason != undefined) {
        self.reportEl.querySelector('.reason').textContent = e.target.title;
      }
    });
    this.reportEl.addEventListener('mouseout', e => {
      self.reportEl.querySelector('.reason').textContent = "";
    });
    modal.querySelector("#modal-policy form").addEventListener('submit', e => {
      this.submitPolicyModal(e);
    });
    if (document.body.classList.contains("mod")) {
      modal.querySelector(".timeout form").addEventListener('submit', e => {
        e.preventDefault();
        self.submitTimeoutModal(e);
        self.hideTimeoutModal();
      });
      modal.querySelector(".timeout a.cancel").addEventListener('click', e => {
        e.preventDefault();
        self.hideTimeoutModal();
      });
      modal.querySelector(".report-clear form").addEventListener('submit', e => {
        e.preventDefault();
        self.submitReportClearModal(e);
        self.hideReportClearModal();
      });
      modal.querySelector(".report-clear a.cancel").addEventListener('click', e => {
        e.preventDefault();
        self.hideReportClearModal();
      });
    }
    this.seriesEl = document.getElementById("series");
    this.seriesEl.addEventListener('click', e => {
      var t = e.target;
      if (t.nodeName == "IMG") {
        t = t.parentNode;
      }
      if (t.nodeName == "A" && t.classList.contains('board')) {
        e.preventDefault();
        game.getSocket().changeBoard(parseInt(t.dataset.id), (board) => {
          Game.setHash();
        });
      }
    });
  };

  nav.prototype.toggleHelp = function(){
    this.toggles.help.click();
  };

  nav.prototype.toggleSeries = function(){
    this.toggles.series.click();
  };

  nav.prototype.toggleMod = function(){
    if ("mod" in this.toggles) {
      this.toggles.mod.click();
    }
  };

  nav.prototype.toggleDemoMode = function(){
    this.demoMode = !this.demoMode;
    localStorage.setItem('demo-mode', this.demoMode);
    window.clearInterval(this.demoInterval);
    const healthbar = document.getElementById("healthbar");
    if (this.demoMode) {
      if (!this.seriesEl.classList.contains('active')) {
        this.toggleSeries();
      }
      this.demoState = {
        boards: [],
        boardNum: 0,
        pause: 0,
      };
      for (let s of Game.Series.list()) {
        for (let b of s.boards) {
          if (!b.active) continue;
          this.demoState.boards.push(b);
        }
      }
      this.demoState.boardNum = this.demoState.boards.length;
      // Toggle leaderboard (not yet exist)
      document.body.classList.add("demo");
      this.demoInterval = setInterval(() => {
        this.demoTick();
      }, 1000);
      this.demoTick();
      healthbar.innerHTML = document.getElementById("tpl-welcome").innerHTML;
      healthbar.style.display = "block";
    } else {
      healthbar.innerHTML = ``;
      healthbar.style.display = "none";
      this.initUser(Game.user());
      document.body.classList.remove("demo");
    }
    return this.demoMode;
  };

  nav.prototype.demoTick = async function(){
    var userEl = document.getElementById('user');
    const user = await Game.User.findLatest();
    if (userEl.dataset.id != user.id) {
      userEl.outerHTML = `
        <a id="user" href="/logout" data-userid="${user.id}" data-policy="true">
            <img src="/u/i/${user.id}"/>
            <span>${user.display_name}</span>
        </a>`
    }
    if (!Game.board()) {
      return;
    }
    if (Game.board().drawnOffset == Game.board().frames.length) {
      this.demoState.pause--;
    }
    if (this.demoState.pause < 1) {
      this.demoState.boardNum++;
      if (this.demoState.boardNum > this.demoState.boards.length - 1) this.demoState.boardNum = 0;
      Game.getSocket().changeBoard(this.demoState.boards[this.demoState.boardNum].id);
      this.demoState.pause = 10;
    }

    // Detect board end reached
    // Detect pause timeout or not started
      // Switch board
  };

  nav.prototype.updateScrubber = function(size) {
    this.scrubber.firstChild.style.width = this.scrubber.offsetWidth + size;
  };

  nav.prototype.resetScrubber = function() {
    this.scrubber.scrollLeft = 0;
  };

  nav.prototype.toggleRecentFrames = function(){
    this.toggles["recent-frames"].click();
  };

  nav.prototype.toggleChat = function(){
    this.toggles["chat"].click();
  };

  nav.prototype.untoggleSiblings = function(id){
    this.toggles[id].parentNode.parentNode.querySelectorAll('.active').forEach((el) => {
      el.classList.remove('active');
    })
    return this.toggles[id];
  };

  nav.prototype.showRecent = function(board) {
    window.cancelAnimationFrame(this.showRecentAnimationFrame);
    this.showRecentAnimationFrame = window.requestAnimationFrame(() => this.doShowRecent(board));
  };

  nav.prototype.doShowRecent = async function(board) {
    var userIds = [];
    var frames = [];
    var f;
    this.recentFrames.classList.remove("board");
    this.recentFrames.classList.remove("tile");
    if (board.focused) {
      for (i = board.tile.frames.length-1; i >= 0; i--) {
        f = board.tile.frames[i];
        if (board.frameIdx[f.timecode] >= board.offset) {
          continue;
        }
        frames.push(f);
        if (frames.length == 10) {
          break;
        }
      };
      this.recentFrames.classList.add("tile");
    } else {
      for (i = board.offset-1; i >= 0; i--) {
        frames.push(board.frames[i]);
        if (frames.length == 10) {
          break;
        }
      }
      this.recentFrames.classList.add("board");
    }
    frames.forEach((f, i) => {
      userIds.push(f.userid);
    });
    if (this.recentTiles.length == 0) {
      var html = '';
      const tpl = document.getElementById("recent-frames-li").innerHTML;
      for (var i = 0; i < 10; i++) {
        this.recentTiles.push(new Game.Tile(null, board.palette, 0, 0, 16));
        html += tpl;
      }
      this.recentFrames.querySelector("ul").innerHTML = html;
      this.recentFrames.querySelectorAll("li").forEach((el, i) => {
        el.prepend(this.recentTiles[i].canvas);
      })
    } else {
      for (let t of this.recentTiles) {
        t.palette = board.palette;
      }
    }
    const userID = Game.userID();
    const users = await g.User.findAll(userIds)
    var li = this.recentFrames.querySelectorAll("li");
    var a, img, span, ta;
    for (var i = 0; i < 10; i++) {
      if (frames.length <= i) {
        li[i].style.visibility = "hidden";
        continue;
      }
      li[i].style.visibility = "visible";
      li[i].dataset.timecode = frames[i].timecode;
      li[i].dataset.userid = frames[i].userid;
      li[i].dataset.date = (+frames[i].date/1000).toFixed(0);
      a = li[i].querySelector('a.user');
      a.title = sanitizeHTML(users[i].display_name);
      a.querySelector('img').src = "/u/i/"+userIds[i];
      a.querySelector('span').textContent = sanitizeHTML(users[i].display_name);
      li[i].querySelector('a.love').style.display = frames[i].userid == userID ? "none": "block";
      li[i].querySelector('.timeago').setAttribute('datetime', frames[i].date.toISOString());
      this.recentTiles[i].renderFrameBuffer(frames[i]);
      this.recentTiles[i].canvas.dataset.i = frames[i].ti;
      this.recentTiles[i].canvas.dataset.j = frames[i].tj;
    }
    this.recentFramesTimeago();
  };

  nav.prototype.showMod = function() {
    window.cancelAnimationFrame(this.showRecentAnimationFrame);
    this.showRecentAnimationFrame = window.requestAnimationFrame(() => this.doShowMod());
  };

  nav.prototype.doShowMod = async function() {
    if (!document.body.classList.contains('mod')) return;
    var users = {};
    var userIds = [];
    var targets = {};
    var targetID = 0;
    var boardId = 0;
    var parts = [];
    var boardStore = {};
    var frameBytes = {};
    var f = {};
    await Game.store().reports.iterate((v, k, i) => {
      parts = k.split("-");
      targetID = parseInt(parts[0]);
      if (!(targetID in targets)) {
        targets[targetID] = [];
      }
      // Append tile number
      targets[targetID].push(v);
    });
    for (var k in targets) {
      if (!targets.hasOwnProperty(k)) continue;
      for (let r of targets[k]) {
        frameBytes = await Game.Series.boardStore(r.boardID).getItem(r.timecode.toString(16).padStart(8, 0));
        f = Game.Frame.fromBytes(frameBytes);
        r.tileNum = f.ti*16 + f.tj%16;
      }
    }
    var userReports = [];
    for (var k in targets) {
      if (targets.hasOwnProperty(k)) {
        userReports.push([k, targets[k]]);
      }
    }
    var uids = {};
    userReports.sort((a, b) => a[1].length > b[1].length ? 1 : (a[1].length < b[1].length ? -1 : 0)).slice(0, 10);
    userReports.forEach((a, i) => {
      uids[a[0]] = true;
      for (let r of a[1]) {
        uids[r.userID] = true;
      }
    });
    for (var k in uids) {
      userIds.push(parseInt(k));
    }
    for (let u of await g.User.findAll(userIds)) {
      if (!u) continue;
      users[u.id] = u;
    }
    var r;
    var html = "<ul>";
    var targetName = "";
    for(let a of userReports) {
      r = a[1];
      r.sort((a, b) => a.frameDate > b.frameDate ? 1 : (a.frameDate < b.frameDate ? -1 : 0));
      html += `<li data-target-id="users[r[0]?.targetID]?.id"><h4>${users[r[0]?.targetID]?.display_name} (${r.length})</h4>`;
      for (let r1 of r) {
        html += `<a href="#${r1.boardID}:${r1.tileNum}:0:3:1" title="${r1.reason} by ${users[r[0]?.userID]?.name}">${r1.boardID} × ${r1.tileNum.toString().padStart(3,0)}</a>`;
      }
      html += `</li>`;
    }
    if (userReports.length == 0) {
      html += `<li><p class="empty">No reports</p></li>`
    }
    html += "</ul>";
    this.modEl.innerHTML = html;
  };

  nav.prototype.showSeries = function(series) {
    var html = "";
    for (let s of series) {
      html += `<li><h4>${s.name}</h4>`;
      for (let b of s.boards) {
        html += `<a class="board" data-id="${b.id}"><img src="${b.bg}"/></a>`;
      }
      html += `</li>`;
    }
    this.seriesEl.querySelector("ul").innerHTML = html;
  };

  nav.prototype.showLoginModal = function() {
    this.modal.classList.add("active", "login");
  };

  nav.prototype.hideLoginModal = function() {
    this.modal.classList.remove("active", "login");
  };

  nav.prototype.showPolicyModal = function() {
    this.modal.classList.add("active", "policy");
  };

  nav.prototype.hidePolicyModal = function() {
    this.modal.classList.remove("active", "policy");
  };

  nav.prototype.showTimeoutModal = function(userid, date) {
    this.modal.classList.add("active", "timeout");
    var form = document.querySelector("#modal-timeout form");
    form.dataset.userid = userid;
    form.dataset.date = date;
  };

  nav.prototype.hideTimeoutModal = function() {
    this.modal.classList.remove("active", "timeout");
  };

  nav.prototype.showReportClearModal = function(userid) {
    this.modal.classList.add("active", "report-clear");
    var form = document.querySelector("#modal-report-clear form");
    form.dataset.userid = userid;
  };

  nav.prototype.hideReportClearModal = function() {
    this.modal.classList.remove("active", "report-clear");
  };

  nav.prototype.submitPolicyModal = function(e) {
    if (!document.getElementById("agree").checked) {
      e.preventDefault();
      return;
    }
  };

  nav.prototype.submitTimeoutModal = function(e) {
    const userID = e.target.dataset.userid;
    const date = e.target.dataset.date;
    var duration = e.submitter.value;
    if (e.submitter.name == 'delete') {
      duration = "-1";
    }
    Game.getSocket().userBan(userID, date, duration);
  };

  nav.prototype.submitReportClearModal = function(e) {
    const userID = e.target.dataset.userid;
    Game.getSocket().clearReports(userID);
  };


  nav.prototype.flash = function(type, msg, timeout) {
    this.flashEl.classList.add('active');
    this.flashEl.innerHTML = `<span class="${type}">${msg}</span>`;
    window.clearTimeout(this.flashTimeout);
    this.flashTimeout = setTimeout(() => {
      this.flashEl.classList.remove('active');
    }, timeout)
  };

  nav.prototype.handleEscape = function() {
    if (this.modal.classList.contains("login")) {
      this.hideLoginModal();
      return true;
    }
    if (this.modal.classList.contains("policy")) {
      window.location.href = "/logout";
      return true;
    }
  };

  nav.prototype.initUser = function(user) {
    var userEl = document.getElementById('user');
    if (user) {
      userEl.outerHTML = `
        <a id="user" href="/logout" data-userid="${user.ID}" data-policy="${user.policy}">
            <img src="${user.profile_image_url}"/>
            <span>${user.display_name}</span>
        </a>`
    } else {
      userEl.outerHTML = document.getElementById('tpl-login').innerHTML;
    }
  };

  nav.prototype.init = function(user) {
    if (this.initialized) {
      return;
    }
    this.initUser(user);
    this.initialized = true;
  };

  /*
  nav.prototype.init = function(user, demo) {
    if (this.initialized) {
      return;
    }
    var userEl = document.getElementById('user');

    var el = document.createElement("div");
    if (user) {
      el.innerHTML = `
        <a id="user" href="/logout" data-userid="${user.ID}" data-policy="${user.policy}">
            <img src="${user.profile_image_url}"/>
            <span>${user.display_name}</span>
        </a>`
    } else {
      userEl.href="/login"
      el.innerHTML = `
        <a id="user" href="/login" class="login">
            <span>Log in with Twitch</span>
        </a>`
    }
    document.querySelector("nav.user").appendChild(el.firstElementChild);
    this.initialized = true;
  };
  */

  nav.prototype.showHeart = function(bucket) {
    if (!(bucket instanceof Array) || bucket.length != 4) {
      return
    }
    clearTimeout(this.heartTimeout);
    const size = bucket[0];
    const rate = bucket[1];
    const level = bucket[2];
    const time = bucket[3];
    var html = "";
    for (var i = 0; i < size; i++) {
      html += ` <span class="heart f`+Math.max(Math.min(level-i*4, 4), 0)+`-4"></span> `;
    }
    const healthbar = document.getElementById("healthbar");
    healthbar.innerHTML = html;
    healthbar.style.display = "block";
    var self = this;
    if (level < size * 4) {
      const diff = Math.floor(Date.now() / 1000) - time;
      this.heartTimeout = setTimeout(() => {
        bucket[2]++
        self.showHeart(bucket);
      }, (diff > 0 && diff < rate ? diff : rate) * 1000);
    }

  };

  return nav

})(Game);
