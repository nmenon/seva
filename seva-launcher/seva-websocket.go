/*
Main websocket control logic
*/
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type WebSocketCommand struct {
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
	ExitCode  int      `json:"exit_code"`
	Response  []string `json:"response"`
}

func websocket_controller(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		var command WebSocketCommand
		err = json.Unmarshal(message, &command)
		if err != nil {
			log.Println(err)
			resp := "Ignoring invalid command"
			log.Println(resp)
			command.ExitCode = 1
			command.Response = append(command.Response, resp)
		}

		switch command.Command {
		case "start_app":
			command = start_app(command)
		case "load_app":
			command = load_app(command)
		case "stop_app":
			command = stop_app(command)
		case "get_app":
			command = get_app(command)
		case "is_running":
			command = is_running(command)
		case "save_settings":
			command = save_settings(command)
		}

		json, err := json.Marshal(command)
		if err != nil {
			log.Println("Error serializing response!")
		}
		err = c.WriteMessage(websocket.TextMessage, []byte(json))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
