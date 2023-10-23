package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsApi(c *gin.Context) {
	fmt.Println(c.Request.Host, c.Request.URL.Query())
	c2s, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Upgrade err", err)
	}

	defer c2s.Close()
	clientCh := make(chan string)
	defer close(clientCh)

	go readClient(clientCh, c2s)
	for {
		m := <-clientCh
		fmt.Println(m)
		if string(m) == "done" {
			break
		}
	}
}
func readClient(ch chan string, conn *websocket.Conn) {
	for {
		t, m, err := conn.ReadMessage()
		fmt.Println(t, "消息类型")
		if err != nil {
			return
		}
		ch <- string(m)
		err = conn.WriteMessage(1, m)
		if err != nil {
			return
		}
	}

}

func httpApi(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.Writer.WriteString("Get method")
	}
	if c.Request.Method == "POST" {
		c.Writer.WriteString("Post method")
	}
}

func main() {
	r := gin.Default()
	r.GET("/stream", wsApi)
	r.GET("/", httpApi)
	r.POST("/", httpApi)
	r.Run(":9000")
}
