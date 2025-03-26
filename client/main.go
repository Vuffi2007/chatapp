package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

func initConnection(conn *websocket.Conn) {
	// RSA pub key should be added here
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	pubE := privateKey.PublicKey.E
	pubN := privateKey.PublicKey.N
	pubEstr := strconv.Itoa(pubE)
	pubNstr := pubN.String()

	var username string

	fmt.Print("Choose a username: ")
	fmt.Scan(&username)

	msg := username + "," + pubEstr + "," + pubNstr + ","
	fmt.Println(msg)

	conn.WriteMessage(websocket.TextMessage, []byte(msg))

}

func readMessages(conn *websocket.Conn, wg *sync.WaitGroup) {
	wg.Done()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}
		println(string(message))
	}
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/echo", nil)
	if err != nil {
		return
	}
	defer conn.Close()

	initConnection(conn)

	var wg sync.WaitGroup
	wg.Add(1)

	go readMessages(conn, &wg)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		line := scanner.Text()

		if line == "" {
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			fmt.Println("Write error:", err)
			break
		}
	}
	wg.Wait()
}
