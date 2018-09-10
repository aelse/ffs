package ffs

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn  *websocket.Conn
	flags map[string]FeatureFlag
}

// NewClient returns a Client connected to ws://addr
func NewClient(addr string) (*Client, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:  c,
		flags: make(map[string]FeatureFlag),
	}

	go func() {
		for {
			var f FeatureFlag
			if err := c.ReadJSON(&f); err != nil {
				log.Printf("Error unmarshaling from websocket: %v", err)
				break
			}
			log.Printf("Client updated flag %s to %v", f.Name, f.Value)
			client.flags[f.Name] = f
		}
	}()

	return client, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Bool(name string, defalt bool) bool {
	f, exists := c.flags[name]
	if exists {
		return f.Value
	}
	return defalt
}
