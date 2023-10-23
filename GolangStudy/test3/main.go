package main

import (
	"fmt"
	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("ss")
	sb, _, err := websocket.DefaultDialer.Dial("ws://localhost:8000/stream", nil)
	if err != nil {
		fmt.Println("Dial backend err", err)
		return
	}
	if err := sb.WriteMessage(websocket.PingMessage, nil); err != nil {
		fmt.Println("ping error ", err)
		return
	}
	fmt.Println("done")
}
