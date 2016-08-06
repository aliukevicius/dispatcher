# Dispatcher

Go package for easier work with websockets.

## Install

```
go get http://gopkg.in/aliukevicius/dispatcher.v1
```

## Import

```
import gopkg.in/aliukevicius/dispatcher.v1
```

## Example

Please check the example folder for details.

### Server side
```go

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

```

### Client side

```js

	d.On('newUser', function (data) {
		console.log("New user: ", data); 
	});

	msg = {
		name: "John Doe",
		email: "john@dispatcher.doe"
	}; 

	d.Emit("createUser", msg);

	d.Join("time", function (data) {
		console.log("Time: ", data);
	});

	d.Join("chat", function (data) {
		console.log("Chat message: ", data)
	});

	d.Broadcast("chat", "Awesome chat message!");

```
