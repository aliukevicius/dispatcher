package dispatcher

var ClientJs = []byte(`var Dispatcher = (function () { 

    var conf = {};    

    function Dispatcher(endpoint) {

        conf.conn = new WebSocket(endpoint); 
        conf.onJoin = null;
        conf.onLeve = null;

        conf.handlers = {};        

        conf.conn.onmessage = function (evt) {
            
            var data = JSON.parse(evt.data);

            if (conf.handlers.hasOwnProperty(data.e)) {
                conf.handlers[data.e](data.m);
            }
        };        
    }    

    Dispatcher.prototype.OnConnect = function (cb) {
        conf.conn.onopen = cb;
    };

    Dispatcher.prototype.OnDisconnect = function (cb) {
        conf.conn.onclose = cb;
    };

    Dispatcher.prototype.OnJoin = function (cb) {
        conf.onJoin = cb;
    };  
    
    Dispatcher.prototype.OnLeve = function (cb) {
        conf.onLeve = cb;
    };

    Dispatcher.prototype.Emit = function (event, message) {

        msg = {
            e: event,
            s: false,
            m: message
        };        

        conf.conn.send(JSON.stringify(msg));
    };

    Dispatcher.prototype.EmitTo = function (recipient, event, message) {
        msg = {
            e: "emitTo",
            s: true,
            m: {
                e: event,
                r: recipient,
                m: message
            }
        };        

        conf.conn.send(JSON.stringify(msg));
    };

    Dispatcher.prototype.On = function (event, cb) {
        conf.handlers[event] = cb;
    };

    Dispatcher.prototype.Join = function (room) {
        msg = {
            e: "join",
            s: true,
            m: room
        };        

        conf.conn.send(JSON.stringify(msg));
    };

    Dispatcher.prototype.Leave = function (room) {
        msg = {
            e: "leave",
            s: true,
            m: room
        };        

        conf.conn.send(JSON.stringify(msg));
    };


    Dispatcher.prototype.Disconnect = function () {
        conf.conn.close();
    };

    return Dispatcher;
} ());`)
