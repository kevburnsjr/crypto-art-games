Game.Nav = (function(g){
  "use strict";

  var nav = function(game, left, right, scrubber){
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
      this.game.setTimecode(scrubber.scrollWidth - scrubber.offsetWidth - scrubber.scrollLeft);
    }, { passive: true });
    scrubber.addEventListener('wheel', e => {
      e.stopPropagation();
      game.setTimecode(scrubber.scrollWidth - scrubber.offsetWidth - scrubber.scrollLeft);
    }, { passive: true });
    this.scrubber = scrubber;
    this.showRecentTimeout = null;
    this.recentFrames = right.querySelector("#recent-frames");
  };

  nav.prototype.toggleHelp = function(){
    this.toggles.help.click();
  };

  nav.prototype.updateScrubber = function(timecode) {
    this.scrubber.firstChild.style.width = this.scrubber.offsetWidth + timecode;
  };

  nav.prototype.showRecent = function(board, timecode) {
    clearTimeout(this.showRecentTimeout);
    this.showRecentTimeout = setTimeout(() => {
      var html = '<ul>';
      board.frames.slice(Math.max(timecode - 10, 0), timecode).reverse().forEach(f => {
        html += '<li>'+f.userid+'</li>';
      });
      this.recentFrames.innerHTML = html + '</ul>';
    }, 200);
  };

  return nav

})(Game);
