package util

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// Hub is a helper to handle one to many chat
type Hub struct {
	conns map[string]net.Conn
	lock  sync.RWMutex
}

// NewHub builds a new hub
func NewHub() *Hub {
	return &Hub{conns: make(map[string]net.Conn)}
}

// Register adds a new conn to the Hub
func (h *Hub) Register(conn net.Conn) {
	fmt.Printf("Connected to %s\n", conn.RemoteAddr())
	h.lock.Lock()
	defer h.lock.Unlock()

	h.conns[conn.RemoteAddr().String()] = conn

	go h.readLoop(conn)
}

func (h *Hub) readLoop(conn net.Conn) {
	b := make([]byte, bufSize)
	for {
		n, err := conn.Read(b)
		if err != nil {
			h.unregister(conn)
			return
		}
		fmt.Printf("Got message: %s\n", string(b[:n]))
	}
}

func (h *Hub) unregister(conn net.Conn) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.conns, conn.RemoteAddr().String())
	err := conn.Close()
	if err != nil {
		fmt.Println("Failed to disconnect", conn.RemoteAddr(), err)
	} else {
		fmt.Println("Disconnected ", conn.RemoteAddr())
	}
}

func (h *Hub) broadcast(msg []byte) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	for _, conn := range h.conns {
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Printf("Failed to write message to %s: %v\n", conn.RemoteAddr(), err)
		}
	}
}

// Chat starts the stdin readloop to dispatch messages to the hub
func (h *Hub) Chat() {
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, err := reader.ReadString('\n')
		Check(err)
		if strings.TrimSpace(msg) == "exit" {
			return
		}
		h.broadcast([]byte(msg))
	}
}
