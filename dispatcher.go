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
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
		},
		connections: map[string]*websocket.Conn{},
	}

	return d
}

func (d *Dispatcher) Handle(w http.ResponseWriter, r *http.Request, h *http.Header) (string, error) {

	c, err := d.upgrader.Upgrade(w, r, *h)
	if err != nil {
		return "", err
	}

	id := uuid.NewV4().String()

	d.connections[id] = c

	go d.readMessages(id)

	return id, nil
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
	}
}
