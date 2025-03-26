package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]string)
var mutex sync.Mutex

func main() {
	http.HandleFunc("/echo", handleConnection)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading username:", err)
		return
	}

	parts := strings.Split(string(msg), ",")

	var username string

	for i, part := range parts {
		switch i {
		case 0:
			username = part
		}
	}

	if len(clients) > 1 { // Only two people is allowed to chat at a time
		return
	}

	mutex.Lock()
	clients[conn] = username
	mutex.Unlock()

	fmt.Println(username, "connected")

	broadcastMessage(username + " connected!")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(username, "disconnected")
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			broadcastMessage(username + " disconnected!")
			return
		}

		broadcastMessage(username + ": " + string(msg))
	}
}

func broadcastMessage(message string) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println("Error broadcasting message:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func sendMessage(message string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		fmt.Println("Error broadcasting message:", err)
		conn.Close()
		_, ok := clients[conn]
		if ok {
			delete(clients, conn)
		}
	}
}
