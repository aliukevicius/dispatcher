package dispatcher

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type DispatcherConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
}

type Dispatcher struct {
	config   *DispatcherConfig
	upgrader websocket.Upgrader

	connections map[string]*Conn
	handlers    map[string]map[string]Handler
	rooms       map[string]map[string]*Conn
}

var defaultDispatcherConfig = &DispatcherConfig{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewDispatcher(config *DispatcherConfig) *Dispatcher {

	if config == nil {
		config = defaultDispatcherConfig
	}

	d := &Dispatcher{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
		},
		connections: map[string]*Conn{},
		handlers:    map[string]map[string]Handler{},
		rooms:       map[string]map[string]*Conn{},
	}

	return d
}

//Handle upgrades http connection to websocket connection
func (d *Dispatcher) Handle(w http.ResponseWriter, r *http.Request, h *http.Header) (*Conn, error) {

	c, err := d.upgrader.Upgrade(w, r, *h)
	if err != nil {
		return nil, err
	}

	id := uuid.NewV4().String()

	d.connections[id] = c

	go d.readMessages(id)

	conn := &Conn{
		dispatcher: d,
		conn:       c,
		ID:         id,
	}

	d.handlers[id] = map[string]Handler{}

	return conn, nil
}

func (d *Dispatcher) readMessages(connectionID string) {

	c := d.connections[connectionID]

	for {
		mt, message, err := c.conn.ReadMessage()
		_ = mt
		_ = message
		if err != nil {
			log.Println("read:", err)
			break
		}
	}

	//close connection if message reading stoped
	d.Close(connectionID)
}

//Broadcast sends message to all connections which are in the room
func (d *Dispatcher) Broadcast(room string, event string, msg interface{}) error {

	connections, ok := d.rooms[room]
	if ok == false {
		return errors.New("Room doesn't exist")
	}

	for _, c := range connections {
		err := c.Emit(event, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

//Close connection
func (d *Dispatcher) Close(ConnectionID string) {

	if c, ok := d.connections[ConnectionID]; ok {
		c.conn.Close()
		delete(d.connections, ConnectionID)
		delete(d.handlers, ConnectionID)
	}
}
