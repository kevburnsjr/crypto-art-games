Game.Nav = (function(g){
  "use strict";

  var nav = function(game, left, right, scrubber, modal){
    this.game = game;
    this.toggles = {};
    const toggleFunc = el => {
      var id = el.dataset.toggle;
      this.toggles[id] = el;
      el.addEventListener("click", (e) => {
        e.preventDefault();
        el.classList.toggle('active');
        document.getElementById(id).classList.toggle('active');
        localforage.setItem("ui-"+id, el.classList.contains('active'));
      });
      localforage.getItem("ui-"+id).then(active => {
        if (active == null) {
          localforage.setItem("ui-"+id, true);
          active = true
        }
        if (active) {
          el.click();
        }
      });
    };
    left.querySelectorAll("nav a").forEach(toggleFunc);
    right.querySelectorAll("nav a").forEach(toggleFunc);
    var interval;
    scrubber.addEventListener('scroll', e => {
      e.stopPropagation();
      if (game.board().tile.active) {
        e.preventDefault();
        return;
      }
      this.game.setTimecode(scrubber.scrollWidth - scrubber.offsetWidth - scrubber.scrollLeft);
    });
    this.scrubber = scrubber;
    this.modal = modal;
    this.showRecentAnimationFrame = null;
    this.recentFrames = right.querySelector("#recent-frames");
    this.recentFrames.addEventListener('wheel', e => {
      e.stopPropagation();
      if (game.board().tile.active) {
        return;
      }
      scrubber.scrollLeft += e.deltaY/Math.abs(e.deltaY);
    }, { passive: true });
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

  nav.prototype.resetScrubber = function(timecode) {
    this.scrubber.scrollLeft = 0;
  };

  nav.prototype.toggleRecentFrames = function(){
    this.toggles["recent-frames"].click();
  };

  nav.prototype.showRecent = function(board) {
    window.cancelAnimationFrame(this.showRecentAnimationFrame);
    this.showRecentAnimationFrame = window.requestAnimationFrame(() => {
      var userIds = [];
      const frames = board.frames.slice(Math.max(board.timecode - 10, 0), Math.max(board.timecode, 1)).reverse();
      frames.forEach((f, i) => {
        userIds.push(f.userid);
      });
      g.User.findAll(userIds).then(users => {
        var html = 'Recent Board Edits<br/><hr/><ul>';
        var tiles = [];
        users.forEach((u, i) => {
          tiles[i] = new Game.Tile(null, board.palette, 0, 0);
          tiles[i].renderFrameBuffer(frames[i]);
          if (!u) {
            html += '<li><a>'+frames[i].userid.toString(16).padStart(4,0)+'</a></li>';
          } else {
            html += '<li><a>';
            if (u.profile_image_url.length > 0) {
              html += '<img src="'+u.profile_image_url+'"/>'
            }
            html += u.display_name+'</a></li>';
          }
        });
        this.recentFrames.innerHTML = html + '</ul>';
        this.recentFrames.querySelectorAll("li").forEach((el, i) => {
          el.prepend(tiles[i].canvas);
        })
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

  return nav

})(Game);
