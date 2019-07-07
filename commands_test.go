package xrpl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type cmdTester interface {
	Test(t *testing.T, mt int, msg []byte) RPCCommand
}

func commandTester(tester cmdTester, t *testing.T) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			tested := tester.Test(t, mt, message)
			err = c.WriteMessage(mt, tested.JSON())
			if err != nil {
				break
			}
		}
	})
}

type peerTester struct{}

func (pt *peerTester) Test(t *testing.T, mt int, msg []byte) RPCCommand {
	cmd := &PeerCommand{}
	err := json.Unmarshal(msg, cmd)
	if err != nil {
		t.Fatal(err)
	}

	return cmd
}

type tickerTester struct {
	received int
}

func (tt *tickerTester) Test(t *testing.T, mt int, msg []byte) RPCCommand {
	cmd := &PeerCommand{}
	err := json.Unmarshal(msg, cmd)
	if err != nil {
		t.Fatal(err)
	}

	tt.received++

	return cmd
}

func TestPeerCommand(t *testing.T) {
	tester := &peerTester{}
	s := httptest.NewServer(commandTester(tester, t))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()
	peerCmd := NewPeerCommand()
	if err := ws.WriteMessage(websocket.TextMessage, peerCmd.JSON()); err != nil {
		t.Fatalf("%v", err)
	}

	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p) != string(peerCmd.JSON()) {
		t.Fatalf("bad message")
	}
}

func TestRepeatCommandTicker(t *testing.T) {
	tester := &tickerTester{received: 0}
	s := httptest.NewServer(commandTester(tester, t))
	defer s.Close()

	node := &Node{Addr: strings.TrimPrefix(s.URL, "http://")}

	peerCmd := NewPeerCommand()

	ctx, cancel := context.WithCancel(context.Background())

	timeout := time.After(4500 * time.Millisecond)
	messages := node.RepeatCommand(ctx, peerCmd, 1)

	for {
		select {
		case <-timeout:
			cancel()
			if tester.received != 4 {
				t.Fatalf("expected 4 ticks, got %d", tester.received)
			}
			return
		case m := <-messages:
			if m != nil {
				t.Log(string(m.Msg))
			}
		}
	}
}
