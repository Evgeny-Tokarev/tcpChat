package server

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"tcpChat/internal/contract"
)

type hub struct {
	Broadcast  chan *contract.BytesMessage
	Register   chan *client
	Unregister chan *client
	Clients    map[*client]struct{}
}

type client struct {
	conn net.Conn
	send chan *contract.BytesMessage
}

func newHub() *hub {
	return &hub{
		Broadcast:  make(chan *contract.BytesMessage, 512),
		Register:   make(chan *client, 16),
		Unregister: make(chan *client, 16),
		Clients:    make(map[*client]struct{}),
	}
}

func newClient(conn net.Conn) *client {
	return &client{
		conn: conn,
		send: make(chan *contract.BytesMessage, 32),
	}
}

func (h *hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.Register:
			h.Clients[client] = struct{}{}
			go h.readWorker(ctx, client)
			go h.writeWorker(ctx, client)
		case client := <-h.Unregister:
			delete(h.Clients, client)
			client.conn.Close()
			close(client.send)
		case message := <-h.Broadcast:
			for client := range h.Clients {
				client.send <- message
			}
		}

	}
}

func (h *hub) readWorker(ctx context.Context, c *client) {
	go func() {
		<-ctx.Done()
		h.Unregister <- c
	}()
	reader := bufio.NewReader(c.conn)
	for {
		bytes, err := reader.ReadBytes(byte('\r'))
		if err != nil {
			if err != io.EOF {
				log.Println("failed to read data, err: ", err)
			}
			return
		}
		log.Printf("read message %s", bytes)
		message, err := contract.NewBytesMessage(bytes)
		if err != nil {
			log.Println(err)
			continue
		}
		h.Broadcast <- message
	}
}

func (h *hub) writeWorker(ctx context.Context, c *client) {
	for {
		select {
		case <-ctx.Done():
		case msg := <-c.send:
			_, err := c.conn.Write(append(msg.Bytes(), '\r'))
			if err != nil {
				log.Println("error c.conn.Write(): ", err)
				return
			}
			log.Printf("send message %s %s", c.conn.LocalAddr().Network(), msg)

		}
	}
}
