package main

import (
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/aliukevicius/dispatcher"
)

type newUser struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Passowrd string `json:"password"`
}

func main() {
	disp := dispatcher.NewDispatcher(nil)

	http.HandleFunc("/", index)
	http.HandleFunc("/ws", webSocket(disp))
	http.HandleFunc("/dispatcher.js", dispatcher.DispatcherJs)

	go timeBroadcaster(disp)

	log.Println("Runing...")
	http.ListenAndServe(":8080", nil)

}

func webSocket(dis *dispatcher.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := dis.Handle(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}

		conn.On("createUser", createUserHandler())
	}
}

func createUserHandler() dispatcher.Handler {
	return func(conn *dispatcher.Conn, data interface{}) {

		uData := data.(map[string]interface{})

		user := newUser{
			ID:       101,
			Name:     uData["name"].(string),
			Email:    uData["email"].(string),
			Passowrd: "password",
		}

		conn.Emit("newUser", user)
	}
}

func timeBroadcaster(disp *dispatcher.Dispatcher) {
	for {
		time.Sleep(time.Second)
		disp.Broadcast("time", time.Now().Format("2006-01-02 15:04:05"))
	}
}

func index(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").ParseFiles("./index.html")
	if err != nil {
		log.Println(err)
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
	}
}
