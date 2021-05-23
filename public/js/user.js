Game.User = (function(g){
  "use strict";

  var user = function(dto){
    this.id = dto.id;
    this.login = dto.login;
    this.display_name = dto.display_name;
    this.profile_image_url = dto.profile_image_url;
  };

  user.prototype.save = async function() {
    return g.store().user.setItem(this.id.toString(16).padStart(4, 0), JSON.stringify(this));
  };

  user.find = async function(userID) {
    return g.store().user.getItem(userID.toString(16).padStart(4, 0)).then(data => {
      return data == null ? null : new user(JSON.parse(data));
    });
  };

  user.findAll = async function(userIDs) {
    var promises = [];
    var users = [];
    userIDs.forEach((id, i) => {
      promises.push(user.find(id).then(u => users[i] = u));
    });
    return Promise.all(promises).then(() => users);
  };

  return user;

})(Game);
