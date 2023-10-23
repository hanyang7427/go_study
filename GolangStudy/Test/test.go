package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsApi(c *gin.Context) {
	fmt.Println(
		"inwsapi",
		"Path-> ", c.Param("path"),
		"appid-> ", c.Param("appid"),
		"userid-> ", c.Param("userid"),
	)
	// upgrade
	h1 := http.Header{
		"Sec-Websocket-Protocol": []string{"streamlit"},
		"Vary":                   []string{"Accept-Encoding"},
		//"Date":                   []string{"Wed, 11 Jan 2023 06:59:35 GMT"},
	}
	cs, err := upGrader.Upgrade(c.Writer, c.Request, h1)
	if err != nil {
		fmt.Println("Upgrade err", err)
	}

	// dial backend
	h2 := http.Header{
		"X-Forwarded-For": []string{c.Request.Host},
		"Host":            []string{"localhost:8501"},
	}
	sb, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("ws://%s:8501/stream", c.Param("appid")),
		h2)
	if err != nil {
		fmt.Println("Dial backend err", err)
		return
	}
	defer sb.Close()
	defer cs.Close()

	csCh := make(chan string)
	bsCh := make(chan string)
	defer close(csCh)
	//转发请求
	go readcs(csCh, cs, sb)
	go readbs(bsCh, sb, cs)
	// cs消息显示
	for {
		select {
		case csmsg := <-csCh:
			_ = csmsg
		case bsmsg := <-bsCh:
			_ = bsmsg
		}
	}
}

func readcs(csCh chan string, cs, sb *websocket.Conn) {
	for {
		t, m, err := cs.ReadMessage()
		if err != nil {
			fmt.Println("Read message error ", err, t)
			return
		}
		fmt.Println("收到client to server 消息", string(m), t)

		csCh <- string(m)

		err = sb.WriteMessage(t, m)
		if err != nil {
			fmt.Println("发送server to backend消息错误", err)
			return
		}
	}
}

func readbs(bsCh chan string, sb, cs *websocket.Conn) {
	for {
		t, m, err := sb.ReadMessage()
		if err != nil {
			fmt.Println("接受backend to server 消息错误", err)
			return
		}
		bsCh <- string(m)

		err = cs.WriteMessage(t, m)
		if err != nil {
			fmt.Println("发送server to client消息错误", err)
			return
		}
	}
}

func httpApi(c *gin.Context) {
	fmt.Println(
		"http wsapi",
		"path-> ", c.Param("path"),
		"userid-> ", c.Param("userid"),
		"appid-> ", c.Param("appid"),
		"host-> ", c.Request.Host,
		"Query-> ", c.Request.URL.Query(),
		"method-> ", c.Request.Method,
		"RequestURI-> ", c.Request.RequestURI,
		"BODY-> ", c.Request.Body)
	// call backend
	client := &http.Client{}
	req, err := http.NewRequest(
		c.Request.Method,
		fmt.Sprintf("http://%s:8501%s", c.Param("appid"), c.Param("path")),
		c.Request.Body,
	)
	if err != nil {
		fmt.Println("make NewRequest error", err)
		return
	}
	req.Header = c.Request.Header
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Do request error", err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Read res error", err)
		return
	}

	for h := range res.Header {
		c.Header(h, res.Header.Get(h))
	}
	c.Data(res.StatusCode, res.Header.Get("Content-Type"), body)
}

func ro(c *gin.Context) {
	param := c.Param("path")
	if param == "/stream" {
		wsApi(c)
	} else {
		httpApi(c)
	}
}

func main() {
	r := gin.Default()
	r.Handle("GET", "/:userid/:appid/*path", ro)
	r.Handle("POST", "/:userid/:appid/*path", ro)
	r.Run(":8000")
}
