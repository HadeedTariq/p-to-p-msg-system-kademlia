package algo

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// ~ let say this is for storing messages in memory
type Messages struct {
	SenderId       NodeID
	ReceiverId     NodeID
	MessageContent string
	MessageId      string
}
type MessagingPeer struct {
	ID           NodeID
	Messages     []Messages
	Port         int
	Address      string
	RoutingTable *RoutingTable
}

func (mp *MessagingPeer) StartServer() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", mp.Port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer ln.Close()

	fmt.Println("Listening on", mp.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go mp.handleConnection(conn)
	}
}

func (mp *MessagingPeer) handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	// Deserialize the message (can use JSON, gob, etc.)
	var msg Messages
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	fmt.Printf("Received message from %s: %s\n", msg.SenderId.String(), msg.MessageContent)

	// Save the message
	mp.Messages = append(mp.Messages, msg)
}

func (mp *MessagingPeer) SendMessage(content string, peerId NodeID) {
	// ~ so may be first I have to check that the peer exist or not

}

func NewMessagingPeer(port int, address string) *MessagingPeer {
	add := fmt.Sprintf("%s:%d", address, port)
	selfID := NewNodeId([]byte(add))
	return &MessagingPeer{
		ID:           selfID,
		Messages:     make([]Messages, 0),
		Port:         port,
		Address:      address,
		RoutingTable: NewRoutingTable(selfID),
	}
}
