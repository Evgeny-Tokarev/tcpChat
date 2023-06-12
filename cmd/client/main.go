package main

import (
	"flag"
	"fmt"
	"log"
	"tcpChat/internal/client"
)

var (
	nickname string
	addr     string
)

func init() {
	flag.StringVar(&nickname, "nickname", "user-xxxx", "client nickname")
	flag.StringVar(&addr, "addr", "0.0.0.0:13003", "server address")
}

func main() {
	flag.Parse()
	fmt.Println("initing: ", nickname, addr)

	cfg := client.Config{
		Addr:     addr,
		Nickname: nickname,
	}

	if err := client.New(cfg).Run(); err != nil {
		log.Fatal(err)
	}
}
