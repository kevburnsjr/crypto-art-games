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
    this.flashEl = document.getElementById("flash");
    this.flashTimeout = null;
    this.heartTimeout = null;
    this.recentTiles = [];
    const toggleFunc = el => {
      var id = el.dataset.toggle;
      this.toggles[id] = el;
      el.addEventListener("click", (e) => {
        e.preventDefault();
        el.classList.toggle('active');
        document.getElementById(id).classList.toggle('active');
        uiStore.setItem("ui-"+id, el.classList.contains('active'));
      });
      uiStore.getItem("ui-"+id).then(active => {
        if (active == null) {
          uiStore.setItem("ui-"+id, true);
          active = true
        }
        if (active) {
          el.click();
        }
      });
    };
    left.querySelectorAll("nav a").forEach(toggleFunc);
    right.querySelectorAll("nav a").forEach(toggleFunc);
    bot.querySelectorAll("nav a").forEach(toggleFunc);
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
      this.game.setTimecode(scrubber.scrollWidth - scrubber.offsetWidth - Math.round(scrubber.scrollLeft*window.devicePixelRatio));
    });
    this.recentFrames.addEventListener('wheel', this.handleWheel, { passive: true });
    this.recentFrames.addEventListener('click', e => {
      e.preventDefault();
      if (e.target.nodeName == "CANVAS") {
        game.board().setFocus(parseInt(e.target.dataset.i), parseInt(e.target.dataset.j));
      }
    });
    var self = this;
    this.recentFrames.addEventListener('mousedown', e => {
      e.preventDefault();
      if (e.target.nodeName == "A" && e.target.classList.contains("report")) {
        self.reportEl.style.right = 0;
        self.reportEl.style.top = e.target.getBoundingClientRect().bottom;
        self.reportEl.dataset.timecode = e.target.parentNode.parentNode.dataset.timecode;
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
        game.getSocket().report(
          e.target.parentNode.dataset.timecode,
          e.target.dataset.reason
        );
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
  };

  nav.prototype.toggleHelp = function(){
    this.toggles.help.click();
  };

  nav.prototype.updateScrubber = function(timecode) {
    this.scrubber.firstChild.style.width = this.scrubber.offsetWidth + timecode;
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

  nav.prototype.doShowRecent = function(board) {
    var userIds = [];
    var frames = [];
    if (board.focused) {
      for (var i = board.tile.frames.length-1; i >= 0; i--) {
        if (board.tile.frames[i].timecode > board.timecode) {
          continue;
        }
        frames.push(board.tile.frames[i]);
        if (frames.length > 10) {
          break;
        }
      }
      this.recentFrames.querySelector('h4').textContent = "Recent Tile Edits";
    } else {
      frames = board.frames.slice(Math.max(board.timecode - 10, 0), Math.max(board.timecode, 0)).reverse();
      this.recentFrames.querySelector('h4').textContent = "Recent Board Edits";
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
    }
    g.User.findAll(userIds).then(users => {
      var li = this.recentFrames.querySelectorAll("li");
      var a, img, span, ta;
      for (var i = 0; i < 10; i++) {
        if (frames.length <= i) {
          li[i].style.visibility = "hidden";
          continue;
        }
        li[i].style.visibility = "visible";
        li[i].dataset.timecode = frames[i].timecode;
        a = li[i].querySelector('a.user');
        img = a.querySelector('img');
        a.title = sanitizeHTML(users[i].display_name);
        if (users[i].profile_image_url.length > 0) {
          img.src = users[i].profile_image_url;
          img.style.display = "block";
        } else {
          img.style.display = "none";
        }
        a.querySelector('span').textContent = sanitizeHTML(users[i].display_name);
        li[i].querySelector('.timeago').setAttribute('datetime', frames[i].date.toISOString());
        this.recentTiles[i].renderFrameBuffer(frames[i]);
        this.recentTiles[i].canvas.dataset.i = frames[i].ti;
        this.recentTiles[i].canvas.dataset.j = frames[i].tj;
      }
      this.recentFramesTimeago();
    });
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

  nav.prototype.submitPolicyModal = function(e) {
    if (!document.getElementById("agree").checked) {
      e.preventDefault();
      return;
    }
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

  nav.prototype.init = function(user) {
    if (this.initialized) {
      return;
    }
    var el = document.createElement("div");
    if (user) {
      el.innerHTML = `
        <a id="user" href="/logout" data-userid="${user.ID}" data-policy="${user.policy}">
            <img src="${user.profile_image_url}"/>
            <span>${user.display_name}</span>
        </a>`
    } else {
      el.innerHTML = `
        <a href="/login" class="login">
            <span>Log in with Twitch</span>
        </a>`
    }
    document.querySelector("nav.user").appendChild(el.firstElementChild);
    this.initialized = true;
  };

  nav.prototype.showHeart = function(bucket) {
    if (!bucket instanceof Array || bucket.length != 4) {
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
