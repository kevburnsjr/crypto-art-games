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
      var failures = 1;
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
          connected = true;
          self.connection.onclose = function(evt) {
            failures++;
            connected = false;
          }
          self.connection.onmessage = function(evt) {
            g.app.online = true;
            failures = Math.min(failures, 10);
            if(evt.ts && typeof evt.ts === 'number') {
              evt.ts = '' + evt.ts;
            }
            try {
              var data = JSON.parse(evt.data);
              self.emit('message', data);
            } catch(e) {
              g.app.log('Socket event parse failed: ' + evt);
              g.app.log(e);
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
