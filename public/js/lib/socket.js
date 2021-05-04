Game.socket = (function (g) {
  "use strict";

  var socket = function(data) {
    var self = this;
    g.object.extend(this, {
      timeout:    null,
      connect:    null,
      connection: null
    });
    g.object.extend(this, data);

    this.stop = function() {
      self.connect = false;
      clearTimeout(self.timeout);
      if(self.connection) {
        self.connection.close();
      }
      self.emit('stop');
    };

    this.start = function() {
      var failures = 0;
      var backoff = function() {
        return Math.max(failures*2*1000, 64 * 1000);
      }
      self.emit('start');
      self.connect = true;
      var connected = false;
      var wrapperfunc = function(){
        if (typeof(WebSocket) === "function" && (!self.connection || self.connection.readyState > 0) && !connected) {
          var uri = new Uri(window.location);
          var host = uri.host();
          var scheme = uri.protocol() == 'https' ? 'wss' : 'ws';
          var port = uri.port() ? ':' + uri.port() : '';
          var url = self.url();
          if(!url) {
            return;
          }
          self.connection = new WebSocket(scheme+"://"+host+port+url);
          self.connection.binaryType = "arraybuffer";
          self.connection.onclose = function(evt) {
            g.online = false;
            failures++;
            connected = false;
          }
          self.connection.onopen = function(evt) {
            g.online = true;
            failures = 0;
            connected = true;
          }
          self.connection.onmessage = function(evt) {
            console.log(evt);
            try {
              self.emit('message', evt.data);
            } catch(e) {
              g.log('Socket event parse failed: ' + evt);
              g.log(e);
            }
          }
        }
        self.timeout = setTimeout(wrapperfunc, backoff());
      };
      wrapperfunc();
    };

    g.event.extend(this);
  };

  return socket;

})(Game);
