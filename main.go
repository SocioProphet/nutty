package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/skyrings/skyring-common/tools/uuid"
)

type termData struct {
	Data string `json:"data"`
	Rows int    `json:"rows"`
	Cols int    `json:"cols"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	uuid, err := uuid.New()
	if err != nil {
		log.Fatal(err)
	}
	u := url.URL{Scheme: "wss", Host: "nutty.io", Path: "/primary/" + uuid.String()}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()
	fmt.Println("https://nutty.io/share/" + uuid.String())
	c := exec.Command("bash")
	ptmx, err := pty.Start(c)
	if err != nil {
		return
	}
	pty.Setsize(ptmx, &pty.Winsize{Rows: 30, Cols: 100})
	go func() {
		for {
			var data termData
			err := conn.ReadJSON(&data)
			if err != nil {
				log.Println(err)
				return
			}
			if _, err := ptmx.Write([]byte(data.Data)); err != nil {
				log.Println(err)
				return
			}
		}
	}()
	for {
		buf := make([]byte, 1024)
		n, err := ptmx.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		if err = conn.WriteJSON(termData{Data: string(buf[:n])}); err != nil {
			log.Println(err)
			return
		}
	}
}
