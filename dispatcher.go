package dispatcher

import (
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

	connections map[string]*websocket.Conn
	handlers    map[string]map[string]Handler
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
		connections: map[string]*websocket.Conn{},
		handlers:    map[string]map[string]Handler{},
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
		mt, message, err := c.ReadMessage()
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

//Close connection
func (d *Dispatcher) Close(ConnectionID string) {

	if c, ok := d.connections[ConnectionID]; ok {
		c.Close()
		delete(d.connections, ConnectionID)
		delete(d.handlers, ConnectionID)
	}
}
