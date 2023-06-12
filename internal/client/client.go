package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/rivo/tview"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"tcpChat/internal/contract"
)

type Client struct {
	cfg Config

	conn net.Conn

	outgoingMessage chan *contract.Message
	incomingMessage chan *contract.Message
}

func New(cfg Config) *Client {
	return &Client{
		cfg:             cfg,
		outgoingMessage: make(chan *contract.Message, 16),
		incomingMessage: make(chan *contract.Message, 0),
	}
}

func (c *Client) Run() error {
	conn, err := net.Dial("tcp", c.cfg.Addr)
	if err != nil {
		return err
	}
	c.conn = conn

	ctx, cancel := context.WithCancel(context.Background())

	go c.tcpRun(ctx)
	go c.uiRun(ctx)

	return gracefulShutdown(cancel)
}

func (c *Client) uiRun(ctx context.Context) {
	app := tview.NewApplication()
	textArea := tview.NewTextArea().
		SetPlaceholder("Enter text here...")
	position := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	messages := ""

	updateInfos := func() {
		s := textArea.GetText()
		if len(s) != 0 && ([]rune(s))[len([]rune(s))-1] == '\n' {
			c.outgoingMessage <- &contract.Message{
				Nick: c.cfg.Nickname,
				Body: s,
			}
			textArea.SetText("", false)
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.incomingMessage:
				messages += fmt.Sprintf("[yellow]%s: [white]%s", msg.Nick, msg.Body)
				position.SetText(messages)
			}
		}
	}()

	textArea.SetMovedFunc(updateInfos)
	updateInfos()

	mainView := tview.NewGrid().
		SetRows(0, 1).
		AddItem(position, 0, 0, 1, 2, 0, 0, false).
		AddItem(textArea, 1, 0, 1, 2, 3, 0, true)

	if err := app.SetRoot(mainView,
		true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
	}
}

func (c *Client) tcpRun(ctx context.Context) {
	go c.messageRead()
	go c.messageWrite(ctx)
	<-ctx.Done()
	c.conn.Close()
}

func (c *Client) messageRead() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\r')
		if len(msg) == 0 {
			continue
		}
		if err != nil {
			log.Println("error while read message, err:", err)
			return
		}
		bmsg, err := contract.NewBytesMessage([]byte(msg))
		if err != nil {
			log.Println("error while creating new message, err:", err)
			continue
		}
		fmt.Println("sending message to c.incomingMessage")
		c.incomingMessage <- bmsg.ToMessage()
	}
}

func (c *Client) messageWrite(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.outgoingMessage:
			fmt.Println("read message from c.outgoingMessage")
			if _, err := c.conn.Write(append(msg.ToBytesMessage().Bytes(), '\r')); err != nil {
				log.Println("error writing message: ", err.Error())
				return
			}
		}
	}
}

func gracefulShutdown(cancel context.CancelFunc) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
	cancel()
	return nil
}
