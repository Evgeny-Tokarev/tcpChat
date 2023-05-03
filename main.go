package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Client struct {
	conn     net.Conn
	username string
}

type ConnectionPool struct {
	clients []*Client
	mutex   sync.Mutex
}

func main() {
	pool := &ConnectionPool{}
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("Server started, listening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		fmt.Println("New client connected:", conn.RemoteAddr().String())

		client := &Client{
			conn: conn,
		}

		pool.AddClient(client)

		go handleClient(pool, client)
	}
}
func (pool *ConnectionPool) AddClient(client *Client) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.clients = append(pool.clients, client)
}

func (pool *ConnectionPool) RemoveClient(client *Client) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for i, c := range pool.clients {
		if c == client {
			pool.clients = append(pool.clients[:i], pool.clients[i+1:]...)
			break
		}
	}
}

func (pool *ConnectionPool) Broadcast(message string, sourceClient *Client) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for _, client := range pool.clients {
		if client.username != sourceClient.username {
			writer := bufio.NewWriter(client.conn)
			_, err := writer.WriteString(message)
			if err != nil {
				fmt.Println(err)
			}
			err = writer.Flush()
			if err != nil {
				return
			}
		}
	}
}
func (pool *ConnectionPool) SandMessage(message string, client *Client) {
	writer := bufio.NewWriter(client.conn)
	_, err := writer.WriteString(message)
	if err != nil {
		fmt.Println(err)
	}
	err = writer.Flush()
	if err != nil {
		return
	}
}
func handleClient(pool *ConnectionPool, client *Client) {
	defer func() {
		pool.RemoveClient(client)
		err := client.conn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	scanner := bufio.NewScanner(client.conn)
	writer := bufio.NewWriter(client.conn)

	_, err := writer.WriteString("Please enter a username: ")
	if err != nil {
		fmt.Println(err)
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println(err)
	}

	scanner.Scan()
	client.username = scanner.Text()

	welcomeMessage := fmt.Sprintf("Welcome to the chat room, %s!\n", client.username)
	_, err = writer.WriteString(welcomeMessage)
	if err != nil {
		fmt.Println(err)
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println(err)
	}

	for scanner.Scan() {
		message := scanner.Text()
		if message == "exit" {
			pool.Broadcast(fmt.Sprintf("%s leaved the chat\n", client.username), client)
			pool.SandMessage(fmt.Sprintf("By-by %s\n", client.username), client)
			break
		}
		fmt.Printf("The message from %s was: %s\n", client.username, message)

		pool.Broadcast(fmt.Sprintf("%s: %s\n", client.username, message), client)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from client: %s\n", err.Error())
	}
}
