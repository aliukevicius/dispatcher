package dispatcher

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type DispatcherConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
}

type Dispatcher struct {
	config   *DispatcherConfig
	upgrader websocket.Upgrader
}

var defaultDispatcherConfig = &DispatcherConfig{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewServer(config *DispatcherConfig) *http.Server {

	if config == nil {
		config = defaultDispatcherConfig
	}

	d := &Dispatcher{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
		},
	}

	s := &http.Server{
		Addr:           ":8050",
		Handler:        d,
		ReadTimeout:    time.Second * 30,
		WriteTimeout:   time.Second * 30,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("It Worksssss!"))

}
