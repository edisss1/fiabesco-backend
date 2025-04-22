package ws

import (
	"github.com/gofiber/websocket/v2"
	"log"
	"sync"
)

type Client struct {
	isClosing bool
	mu        sync.Mutex
}

var (
	Clients    = make(map[*websocket.Conn]*Client)
	Register   = make(chan *websocket.Conn)
	Broadcast  = make(chan string)
	Unregister = make(chan *websocket.Conn)
)

func RunHub() {
	for {
		select {
		case conn := <-Register:
			Clients[conn] = &Client{}
		case msg := <-Broadcast:

			log.Printf("Message received: %v", msg)
			for conn, c := range Clients {
				go func(conn *websocket.Conn, c *Client) {
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.isClosing {
						return
					}
					if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
						c.isClosing = true
						log.Println("write error: ", err)

						conn.WriteMessage(websocket.CloseMessage, []byte{})
						conn.Close()
						Unregister <- conn
					}
				}(conn, c)
			}
		case conn := <-Unregister:
			delete(Clients, conn)

			log.Println("Connection unregistered")
		}
	}
}
