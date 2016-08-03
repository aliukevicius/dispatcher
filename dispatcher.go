package dispatcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type MessageData struct {
	Event   string      `json:"e"`
	System  bool        `json:"s"`
	Message interface{} `json:"m"`
}

type BroadcastMessage struct {
	Room    string      `json:"r"`
	Message interface{} `json:"m"`
}

type DispatcherConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
}

type Dispatcher struct {
	config   *DispatcherConfig
	upgrader websocket.Upgrader

	handlers    map[string]map[string]Handler
	rooms       map[string]map[string]*Conn
	connections map[string]*Conn
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
		handlers:    map[string]map[string]Handler{},
		rooms:       map[string]map[string]*Conn{},
		connections: map[string]*Conn{},
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
		rooms:      map[string]map[string]*Conn{},
		lock:       &sync.Mutex{},
	}

	d.connections[id] = conn
	d.handlers[id] = map[string]Handler{}

	go d.readMessages(id)

	return conn, nil
}

func (d *Dispatcher) readMessages(connectionID string) {

	c := d.connections[connectionID]

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
			) == false {
				log.Println("read:", err)
			}

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
			} else if msg.Event == "leave" {

				d.leave(c, msg.Message.(string))
			} else if msg.Event == "emitTo" {

				m := msg.Message.(map[string]interface{})

				recipient := m["r"].(string)
				event := m["e"].(string)

				err := d.EmitTo(recipient, event, m["m"])
				if err != nil {
					log.Println(err)
				}
			} else if msg.Event == "broadcast" {

				m := msg.Message.(map[string]interface{})
				room := m["r"].(string)

				err := d.Broadcast(room, m["m"])
				if err != nil {
					log.Println(err)
				}
			}

			//all the system event handling ends here
			continue
		}

		if handler, ok := d.handlers[c.ID][msg.Event]; ok {
			handler(c, msg.Message)
		}
	}

	//close connection if message reading stoped
	d.Close(connectionID)
}

//EmitTo send message to all connections in the room
func (d *Dispatcher) EmitTo(recipient string, event string, msg interface{}) error {

	conn, ok := d.connections[recipient]
	if ok == false {
		return fmt.Errorf("Recipient '%s' doesn't exist.", recipient)
	}

	err := conn.Emit(event, msg)
	if err != nil {
		return err
	}

	return nil
}

//Broadcast send message to a room subscribers
func (d *Dispatcher) Broadcast(room string, message interface{}) error {

	connections, ok := d.rooms[room]
	if ok == false {
		return fmt.Errorf("Room '%s' doesn't exist.", room)
	}

	msg := MessageData{
		Event:  "broadcast",
		System: true,
		Message: BroadcastMessage{
			Room:    room,
			Message: message,
		},
	}

	for _, conn := range connections {
		err := conn.emit(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

//Close connection
func (d *Dispatcher) Close(ConnectionID string) error {

	conn, ok := d.connections[ConnectionID]
	if ok == false {
		return fmt.Errorf("Connection with '%s' ID not found", ConnectionID)
	}

	// remove connection from all the rooms it belongs to
	for name, room := range conn.rooms {
		if len(room) > 1 {
			// remove connection from the room
			delete(room, ConnectionID)
		} else {
			// delete room
			delete(d.rooms, name)
		}
	}

	delete(d.handlers, ConnectionID)
	delete(d.connections, ConnectionID)
	conn.close()

	return nil
}

func (d *Dispatcher) join(conn *Conn, roomName string) {

	_, ok := d.rooms[roomName]
	if ok == false {
		d.rooms[roomName] = map[string]*Conn{}
	}

	d.rooms[roomName][conn.ID] = conn

	conn.rooms[roomName] = d.rooms[roomName]
}

func (d *Dispatcher) leave(conn *Conn, roomName string) {

	_, ok := d.rooms[roomName]
	if ok == false {
		return // can't find room, so there is nothing to leve
	}

	// check if room has more than one connection
	if len(conn.rooms[roomName]) > 1 {
		// remove connection from the room
		delete(conn.rooms[roomName], conn.ID)
	} else {
		// room has only one connection so we will remove the room
		delete(d.rooms, roomName)

	}

	// remove room data from connection
	delete(conn.rooms, roomName)
}

//DispatcherJs handles client js code requests
func DispatcherJs(w http.ResponseWriter, r *http.Request) {

	w.Header()["Content-Type"] = []string{"application/javascript"}

	w.Write(ClientJs)
}
