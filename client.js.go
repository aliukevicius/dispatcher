package dispatcher

var ClientJs = []byte(`var Dispatcher = (function () { 

    var conf = {
        onJoin: null,
        onLeve: null,
        handlers: {},
        roomHandlers: {},
        conn: null
    };    

    function Dispatcher(endpoint) {

        conf.conn = new WebSocket(endpoint); 

        conf.conn.onmessage = function (msg) {
            var data = JSON.parse(msg.data);

            if (data.s === true) {
                if (data.e == "broadcast") {
                    if (conf.roomHandlers.hasOwnProperty(data.m.r)) {
                        conf.roomHandlers[data.m.r](data.m.m);
                    }
                }
            } else {
                if (conf.handlers.hasOwnProperty(data.e)) {
                    conf.handlers[data.e](data.m);
                }
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

    Dispatcher.prototype.Broadcast = function (room, message) {
        msg = {
            e: "broadcast",
            s: true,
            m: {
                r: room,
                m: message
            }
        };        

        conf.conn.send(JSON.stringify(msg));
    };

    Dispatcher.prototype.On = function (event, cb) {
        conf.handlers[event] = cb;
    };

    Dispatcher.prototype.Join = function (room, cb) {
        conf.roomHandlers[room] = cb;

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
