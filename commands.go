package xrpl

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// Node is an XRPL proxy or validator node
type Node struct {
	Addr string
	Port string
	tls  bool
}

// NewNode creates a new XRPL node representation
func NewNode(addr, port string, tls bool) *Node {
	return &Node{Addr: addr, Port: port, tls: tls}
}

func (n *Node) connect() *websocket.Conn {
	scheme := "ws"
	if n.tls {
		scheme = "wss"
	}

	url := url.URL{Scheme: scheme, Host: net.JoinHostPort(n.Addr, n.Port)}
	log.Printf("websocket connect: %s", url.String())

	c, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal("websocket dial:", err)
	}
	log.Printf("websocket connected...")

	return c
}

// RepeatCommand dials the ws/wss RPC endpoint for admin commands then repeats
// a command periodically
func (n *Node) RepeatCommand(ctx context.Context, cmd RPCCommand, repeat int) chan *WsMessage {
	messages := make(chan *WsMessage)

	c := n.connect()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("reader exiting")
				close(messages)
				return
			default:
				msgType, message, err := c.ReadMessage()
				if err != nil {
					log.Println("websocket read:", err)
				}
				messages <- &WsMessage{MsgType: msgType, Msg: message, Err: err}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(repeat))
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("websocket close")
				err := c.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				if err != nil {
					log.Println("websocket close:", err)
					return
				}
				select {
				case <-time.After(time.Millisecond * 500):
				}
				return
			case <-ticker.C:
				err := c.WriteMessage(websocket.TextMessage, cmd.JSON())
				if err != nil {
					log.Println("websocket write:", err)
					return
				}
			}
		}
	}()

	return messages
}

// DoCommand runs a single command
func (n *Node) DoCommand(cmd RPCCommand) *WsMessage {

	c := n.connect()

	go func() {
		err := c.WriteMessage(websocket.TextMessage, cmd.JSON())
		if err != nil {
			log.Println("websocket write:", err)
			return
		}
	}()

	msgType, message, err := c.ReadMessage()
	if err != nil {
		log.Println("websocket read:", err)
	}

	err = c.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		log.Println("websocket close:", err)
		return nil
	}

	select {
	case <-time.After(time.Millisecond * 200):
	}

	return &WsMessage{MsgType: msgType, Msg: message, Err: err}
}

// WsMessage encapsulates a websocket txt/binary message
type WsMessage struct {
	MsgType int
	Msg     []byte
	Err     error
}

// RPCCommand defines commands that can be sent to the nodes
type RPCCommand interface {
	JSON() []byte
}

// Command is a rippled admin command
type Command struct {
	Command       string `json:"command"`
	AdminUser     string `json:"admin_user,omitempty"`
	AdminPassword string `json:"admin_password,omitempty"`
}

// Name returns the commands string name e.g. "peers"
func (c *Command) Name() string {
	return c.Command
}

// JSON implements the RPCCommand interface in admin commands
func (c *Command) JSON() []byte {
	cmdJSON, err := json.Marshal(c)
	if err != nil {
		log.Fatal("command:", err)
	}

	return cmdJSON
}

// PeerCommand is a "peers" rpc admin command
type PeerCommand struct {
	*Command
}

// NewPeerCommand creates a new "peers" command
func NewPeerCommand() *PeerCommand {
	return &PeerCommand{
		Command: &Command{"peers", "", ""},
	}
}
