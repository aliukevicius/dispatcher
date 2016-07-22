package dispatcher

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Handler func(*websocket.Conn, interface{})

type Conn struct {
	dispatcher *Dispatcher
	conn       *websocket.Conn
	ID         string
	//rooms to which connection belongs to
	rooms map[string]map[string]*Conn
}

//On assigns handler for event
func (c *Conn) On(event string, handler Handler) {

	c.dispatcher.handlers[c.ID][event] = handler
}

//Emit sends message for particular event
func (c *Conn) Emit(event string, message interface{}) error {

	mt := websocket.TextMessage

	m := MessageData{
		Event:   event,
		System:  false,
		Message: message,
	}

	msg, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = c.conn.WriteMessage(mt, msg)
	if err != nil {
		return err
	}

	return nil
}
