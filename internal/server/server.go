package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	hub *hub
}

func New() *Server {
	return &Server{
		hub: newHub(),
	}
}
func (s *Server) Run(cfg Config) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		return err
	}
	fmt.Printf("Server started on port %d\n", cfg.Port)
	ctx, cancel := context.WithCancel(context.Background())
	go s.runListener(ctx, listener)
	go s.hub.Run(ctx)
	return gracefulShutdown(cancel)
}

func gracefulShutdown(cancel context.CancelFunc) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
	cancel()
	return nil
}

func (s *Server) runListener(ctx context.Context, listener net.Listener) {
	go func() {
		<-ctx.Done()
		err := listener.Close()
		if err != nil {
			log.Println("error closing connect")
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept connection, err: ", err)
		}
		log.Println("connect new client ", conn.LocalAddr().Network(), conn.LocalAddr().String())
		s.hub.Register <- newClient(conn)
	}
}
