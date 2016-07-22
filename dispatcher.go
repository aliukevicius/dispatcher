package dispatcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type MessageData struct {
	Event   string      `json:"e"`
	System  bool        `json:"s"`
	Message interface{} `json:"m"`
}

type DispatcherConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
}

type Dispatcher struct {
	config   *DispatcherConfig
	upgrader websocket.Upgrader

	handlers map[string]map[string]Handler
	rooms    map[string]map[string]*Conn
}

var defaultDispatcherConfig = &DispatcherConfig{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//NewDispatcher create dispatcher
func NewDispatcher(config *DispatcherConfig) *Dispatcher {

	if config == nil {
		config = defaultDispatcherConfig
	}

	d := &Dispatcher{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
		},
		handlers: map[string]map[string]Handler{},
		rooms:    map[string]map[string]*Conn{},
	}

	return d
}

//Handle upgrades http connection to websocket connection
func (d *Dispatcher) Handle(w http.ResponseWriter, r *http.Request, h *http.Header) (*Conn, error) {

	var c *websocket.Conn
	var err error

	if h == nil {
		c, err = d.upgrader.Upgrade(w, r, nil)
	} else {
		c, err = d.upgrader.Upgrade(w, r, *h)
	}

	if err != nil {
		return nil, err
	}

	id := uuid.NewV4().String()

	conn := &Conn{
		dispatcher: d,
		conn:       c,
		ID:         id,
	}

	// ecah connection has it's seperate room with a room name as a connection ID
	d.rooms[id] = map[string]*Conn{}
	d.rooms[id][id] = conn

	d.handlers[id] = map[string]Handler{}

	go d.readMessages(id)

	return conn, nil
}

func (d *Dispatcher) readMessages(connectionID string) {

	c := d.rooms[connectionID][connectionID]

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		msg := &MessageData{}
		json.Unmarshal(message, msg)
		if err != nil {
			log.Println("MessageData unmarshal:", err)
			break
		}

		// handle system events
		if msg.System {

			if msg.Event == "join" {
				d.join(c, msg.Message.(string))
			}

			if msg.Event == "leave" {
				d.leave(c, msg.Message.(string))
			}

			//all the system event handling ends here
			continue
		}

		if handler, ok := d.handlers[c.ID][msg.Event]; ok {
			handler(c.conn, msg.Message)
		}
	}

	//close connection if message reading stoped
	d.Close(connectionID)
}

//EmitTo send message to all connections in the room
func (d *Dispatcher) EmitTo(room string, event string, msg interface{}) error {

	connections, ok := d.rooms[room]
	if ok == false {
		return fmt.Errorf("Romm with %s name doesn't exist.", room)
	}

	// send message to each connection in the room
	for _, conn := range connections {
		err := conn.Emit(event, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

//Close connection
func (d *Dispatcher) Close(ConnectionID string) error {

	conn, ok := d.rooms[ConnectionID][ConnectionID]
	if ok == false {
		return fmt.Errorf("Connection with '%s' ID not found", ConnectionID)
	}

	//todo: remove connection from each room to which it belongs to

	delete(d.rooms, ConnectionID)
	delete(d.handlers, ConnectionID)
	conn.conn.Close()

	return nil
}

func (d *Dispatcher) join(conn *Conn, roomName string) {

	_, ok := d.rooms[roomName]
	if ok == false {
		d.rooms[roomName] = map[string]*Conn{}
	}

	d.rooms[roomName][conn.ID] = conn

	//todo: add room name to conneciton object

}

func (d *Dispatcher) leave(conn *Conn, roomName string) {

	_, ok := d.rooms[roomName]
	if ok == false {
		return // can't find room, so there is nothing to leve
	}

	// check if room has more than one connection
	if len(d.rooms[roomName]) > 1 {
		// remove connection from the room
		delete(d.rooms[roomName], conn.ID)
	} else {
		// room has only one connection so we will remove the room
		delete(d.rooms, roomName)
	}

	//todo: remove rooms name from connection
}
