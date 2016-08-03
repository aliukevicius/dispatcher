package dispatcher

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Handler func(*Conn, interface{})

type ConnCloseHandler func(*Conn)

type Conn struct {
	dispatcher *Dispatcher
	conn       *websocket.Conn
	ID         string
	//rooms to which connection belongs to
	rooms        map[string]map[string]*Conn
	closeHandler ConnCloseHandler

	lock *sync.Mutex
}

//On assigns handler for event
func (c *Conn) On(event string, handler Handler) {

	c.dispatcher.handlers[c.ID][event] = handler
}

//Emit sends message for particular event
func (c *Conn) Emit(event string, message interface{}) error {

	m := MessageData{
		Event:   event,
		System:  false,
		Message: message,
	}

	return c.emit(m)
}

func (c *Conn) emit(message MessageData) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	mt := websocket.TextMessage

	err = c.conn.WriteMessage(mt, msg)
	if err != nil {
		return err
	}

	return nil
}

//OnClose subscribe to connection close
func (c *Conn) OnClose(h ConnCloseHandler) {
	c.closeHandler = h
}

func (c *Conn) close() {
	c.conn.Close()

	if c.closeHandler != nil {
		c.closeHandler(c)
	}
}
