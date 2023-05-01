package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/valyala/fastjson"
	"net/http"
)

var ug = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func V2ray(c *gin.Context) {
	conn, err := ug.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	go func() {
		for {
			_, msg, _ := conn.ReadMessage()
			v, _ := fastjson.Parse(string(msg))
			v.get
		}
	}()

}
