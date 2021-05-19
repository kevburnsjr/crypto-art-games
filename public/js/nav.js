Game.Nav = (function(g){
  "use strict";

  var nav = function(left){
    left.querySelectorAll("nav a").forEach(el => {
      var id = el.dataset.toggle;
      el.addEventListener("click", (e) => {
        e.preventDefault();
        el.classList.toggle('active');
        document.getElementById(id).classList.toggle('active');
        localforage.setItem("ui-"+id, e.target.classList.contains('active'));
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
    });
  };

  nav.prototype.toggleHelp = function(){
    document.getElementById("help-toggle").click();
  };

  return nav

})(Game);
