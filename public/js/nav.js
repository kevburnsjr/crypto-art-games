Game.Nav = (function(g){
  "use strict";

  var nav = function(game, uiStore, left, right, scrubber, modal){
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
    scrubber.addEventListener('scroll', e => {
      e.stopPropagation();
      if (game.board().tile.active) {
        e.preventDefault();
        return;
      }
      this.game.setTimecode(scrubber.scrollWidth - scrubber.offsetWidth - scrubber.scrollLeft);
    });
    this.recentFrames.addEventListener('wheel', e => {
      e.stopPropagation();
      if (game.board().tile.active) {
        return;
      }
      scrubber.scrollLeft += e.deltaY/Math.abs(e.deltaY);
    }, { passive: true });
    this.recentFrames.addEventListener('click', e => {
      e.preventDefault();
      if (e.target.nodeName == "CANVAS") {
        console.log(e.target.dataset.i, e.target.dataset.j);
        game.board().setFocus(parseInt(e.target.dataset.i), parseInt(e.target.dataset.j));
      }
    });
    var self = this;
    this.recentFrames.addEventListener('mousedown', e => {
      e.preventDefault();
      if (e.target.nodeName == "A" && e.target.classList.contains("report")) {
        self.reportEl.style.right = 0;
        self.reportEl.style.top = e.target.getBoundingClientRect().bottom;
        self.reportEl.dataset.timecode = e.target.dataset.timecode;
        document.querySelectorAll('.reporting').forEach((el) => el.classList.remove('reporting'));
        document.body.classList.add('reporting');
        e.target.parentNode.parentNode.classList.add('reporting');
      }
    });
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
        self.reportEl.querySelector('.reason').innerHTML = e.target.title;
      }
    });
    this.reportEl.addEventListener('mouseout', e => {
      self.reportEl.querySelector('.reason').innerHTML = "";
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

  nav.prototype.showRecent = function(board) {
    window.cancelAnimationFrame(this.showRecentAnimationFrame);
    this.showRecentAnimationFrame = window.requestAnimationFrame(() => {
      var userIds = [];
      const frames = board.frames.slice(Math.max(board.timecode - 10, 0), Math.max(board.timecode, 0)).reverse();
      frames.forEach((f, i) => {
        userIds.push(f.userid);
      });
      g.User.findAll(userIds).then(users => {
        var html = '';
        var tiles = [];
        users.forEach((u, i) => {
          tiles[i] = new Game.Tile(null, board.palette, 0, 0);
          tiles[i].renderFrameBuffer(frames[i]);
          html += '<li>';
          html += '<nav class="mod">';
          html += '<a class="love" data-timecode="'+frames[i].timecode+'" title="Love"><span class="heart f4-4"/></a>';
          html += '<a class="report" data-timecode="'+frames[i].timecode+'" title="Report"></a>';
          html += '</nav>';
          html += '<a class="user" title="'+sanitizeHTML(u.display_name)+'">';
          if (u === null) {
            html += frames[i].userid.toString(16).padStart(4,0);
          } else {
            if (u.profile_image_url.length > 0) {
              html += '<img src="'+u.profile_image_url+'"/>'
            }
            html += sanitizeHTML(u.display_name);
          }
          html += '</a>';
          html += '<span class="timeago" datetime="'+frames[i].date.toISOString()+'"/>';
          html += '</li>';
        });
        this.recentFrames.querySelector("ul").innerHTML = html;
        this.recentFrames.querySelectorAll("li").forEach((el, i) => {
          tiles[i].canvas.dataset.i = frames[i].ti;
          tiles[i].canvas.dataset.j = frames[i].tj;
          el.prepend(tiles[i].canvas);
        });
        const ta = this.recentFrames.querySelectorAll('.timeago');
        if (ta.length > 0) {
          timeago.render(ta, 'en_US', {minInterval: 60});
        }
      });
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
    clearTimeout(this.heartTimeout);
    var html = ""
    for (var i = 0; i < bucket.size; i++) {
      html += ` <span class="heart f`+Math.max(Math.min(bucket.level-i*4, 4), 0)+`-4"></span> `;
    }
    document.getElementById("healthbar").innerHTML = html;
    var self = this;
    if (bucket.level < bucket.size * 4) {
      this.heartTimeout = setTimeout(() => {
        bucket.level++
        self.showHeart(bucket);
      }, 60000 / bucket.rate);
    }
  };

  return nav

})(Game);
