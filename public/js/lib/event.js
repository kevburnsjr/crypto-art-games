Game.event = (function (t) {
  "use strict";

  function extend(target){
    if(!target) return;
    var events = {};
    var once = {};
    var key = Math.random();
    var serialQueue = [];
    var processing = false;
    target.on = function(types, fn) {
      types = typeof types === "string" ? [types] : types;
      if(Array.isArray(types)) {
        types.forEach(function(type) {
          events[type] = events[type] || [];
          events[type].push(fn);
        });
      }
    };
    target.once = function(type, fn) {
      once[type] = once[type] || [];
      once[type].push(fn);
    };
    target.off = function(type) {
      if(!type) {
        events = {};
        once = {};
        key = Math.random();
      } else {
        events[type] = [];
        once[type] = [];
      }
    };
    target.emit = function(type, arg) {
      var queue = [].concat(events[type] || []).concat(once[type] || []);
      var k = key;
      var p = [];
      queue.forEach(function(fn) {
        if(k != key) {
          return;
        }
        const ret = fn(arg);
        if (ret instanceof Promise) {
          p.push(ret);
        }
      });
      once[type] = [];
      return p;
    };
    target.all = function(type, arg) {
      return Promise.all(target.emit(type, arg));
    };
    target.serial = function(type, msg) {
      serialQueue.push([type, msg]);
      serialProcess();
    };
    var serialProcess = function(loop) {
      if (processing && !loop) {
        return;
      }
      if (serialQueue.length == 0) {
        processing = false;
        return;
      }
      processing = true;
      const e = serialQueue.shift();
      Promise.all(target.emit(e[0], e[1])).then(() => serialProcess(true));
    };
    return target;
  };

  return extend({
    extend: extend
  });

})(Game);
