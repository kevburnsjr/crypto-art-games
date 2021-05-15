Game.socket = (function (g) {
  "use strict";

  var socket = function(data) {
    var self = this;
    g.object.extend(this, {
      timeout:    null,
      connect:    null,
      connected:  false,
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
        return Math.max(failures*1000, 64 * 1000);
      }
      self.emit('start');
      self.connect = true;
      var wrapperfunc = function(){
        if (typeof(WebSocket) === "function" && (!self.connection || self.connection.readyState > 0) && !self.connected) {
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
            self.connected = false;
          }
          self.connection.onopen = function(evt) {
            g.online = true;
            failures = 0;
            self.connected = true;
          }
          self.connection.onmessage = function(evt) {
            try {
              self.emit('message', evt.data);
            } catch(e) {
              g.log('Socket event parse failed', evt);
              g.log(e);
            }
          }
        }
        self.timeout = setTimeout(wrapperfunc, backoff());
      };
      wrapperfunc();
    };

    this.send = function(msg, retries) {
      if (retries > 10) {
        return
      }
      try {
        this.connection.send(msg);
      } catch(e) {
        console.log(e);
        this.connected = false;
        setTimeout(this.send, 200, msg, (retries ? parseInt(retries) : 0) + 1);
      }
    };

    g.event.extend(this);
  };

  return socket;

})(Game);
