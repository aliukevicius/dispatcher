package dispatcher

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type MessageData map[string]interface{}

type Handler func(*websocket.Conn, MessageData)

type Conn struct {
	dispatcher *Dispatcher
	conn       *websocket.Conn
	ID         string
}

//On assigns handler for event
func (c *Conn) On(event string, handler Handler) {

	c.dispatcher.handlers[c.ID][event] = handler
}

//Emit sends message for particular event
func (c *Conn) Emit(event string, message MessageData) error {

	mt := websocket.BinaryMessage

	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = c.conn.WriteMessage(mt, msg)
	if err != nil {
		return err
	}

	return nil
}

//Broadcast sends message to all connections which are in the room
func (c *Conn) Broadcast(room string, event string, msg MessageData) {

}