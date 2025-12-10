package ws

import (
    "encoding/json"
    "time"

    "github.com/gofiber/fiber/v2"
    websocket "github.com/gofiber/websocket/v2"
)

type Client struct {
    hub *Hub
    conn *websocket.Conn
    send chan []byte
}

func ServeWS(hub *Hub) fiber.Handler {
    return websocket.New(func(c *websocket.Conn) {
        client := &Client{hub: hub, conn: c, send: make(chan []byte, 256)}
        client.hub.register <- client
        go client.writePump()
        client.readPump()
    })
}

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    c.conn.SetReadLimit(512)
    _ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        var tmp map[string]interface{}
        _ = json.Unmarshal(message, &tmp)
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
            _ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                _ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }
        case <-ticker.C:
            _ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
